package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/djbook/backend/internal/model"
	"github.com/djbook/backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

// GoogleTokenInfo is the response from Google's tokeninfo endpoint.
type GoogleTokenInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Aud   string `json:"aud"`
	Exp   string `json:"exp"`
}

// VerifyGoogleToken validates a Google ID token and returns the subject (Google user ID) and email.
func VerifyGoogleToken(ctx context.Context, idToken string) (googleID, email string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://oauth2.googleapis.com/tokeninfo?id_token="+idToken, nil)
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("tokeninfo request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("invalid token: %s", string(body))
	}

	var info GoogleTokenInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return "", "", fmt.Errorf("parse response: %w", err)
	}
	if info.Sub == "" || info.Email == "" {
		return "", "", fmt.Errorf("incomplete token info")
	}
	return info.Sub, info.Email, nil
}

// ApplePublicKey represents an Apple JWK.
type ApplePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleKeySet struct {
	Keys []ApplePublicKey `json:"keys"`
}

// VerifyAppleToken validates an Apple identity token and returns the subject (Apple user ID) and email.
func VerifyAppleToken(ctx context.Context, idToken string) (appleID, email string, err error) {
	// Parse the header to get the key ID
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return "", "", fmt.Errorf("parse token header: %w", err)
	}
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return "", "", fmt.Errorf("missing kid in token header")
	}

	// Fetch Apple's public keys
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return "", "", fmt.Errorf("build keys request: %w", err)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetch apple keys: %w", err)
	}
	defer resp.Body.Close()

	var keySet appleKeySet
	if err := json.NewDecoder(resp.Body).Decode(&keySet); err != nil {
		return "", "", fmt.Errorf("decode apple keys: %w", err)
	}

	// Find matching key
	var matchingKey *ApplePublicKey
	for i := range keySet.Keys {
		if keySet.Keys[i].Kid == kid {
			matchingKey = &keySet.Keys[i]
			break
		}
	}
	if matchingKey == nil {
		return "", "", fmt.Errorf("no matching key found for kid: %s", kid)
	}

	// Build RSA public key
	pubKey, err := buildRSAPublicKey(matchingKey)
	if err != nil {
		return "", "", fmt.Errorf("build rsa key: %w", err)
	}

	// Verify token
	parsed, err := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		return "", "", fmt.Errorf("verify apple token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return "", "", fmt.Errorf("invalid token claims")
	}

	sub, _ := claims["sub"].(string)
	emailVal, _ := claims["email"].(string)
	if sub == "" {
		return "", "", fmt.Errorf("missing sub in token")
	}
	return sub, emailVal, nil
}

func buildRSAPublicKey(key *ApplePublicKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("decode N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// Handler serves HTTP auth endpoints.
type Handler struct {
	userSvc      *service.UserService
	profileSvc   *service.ProfileService
	jwtSecret    string
	demoUsername string
	demoPassword string
}

// NewHandler creates an auth HTTP handler.
func NewHandler(
	userSvc *service.UserService,
	profileSvc *service.ProfileService,
	jwtSecret, demoUsername, demoPassword string,
) *Handler {
	return &Handler{
		userSvc:      userSvc,
		profileSvc:   profileSvc,
		jwtSecret:    jwtSecret,
		demoUsername: demoUsername,
		demoPassword: demoPassword,
	}
}

type credentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authProfileResponse struct {
	ID     string  `json:"id"`
	DjName string  `json:"djName"`
	Bio    *string `json:"bio,omitempty"`
}

type authResponse struct {
	Token    string                `json:"token"`
	UserID   string                `json:"userId"`
	Profiles []authProfileResponse `json:"profiles"`
}

// Register handles POST /auth/register with body { "username", "password" }.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(req.Username)
	if len(username) < 3 {
		http.Error(w, "username must be at least 3 characters", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	user, err := h.userSvc.Register(r.Context(), username, req.Password)
	if err != nil {
		if err == service.ErrUsernameTaken {
			http.Error(w, "username already taken", http.StatusConflict)
			return
		}
		http.Error(w, "user error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.profileSvc.Create(r.Context(), user.ID, &model.DjProfile{DjName: username}, []string{})
	if err != nil {
		http.Error(w, "profile error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeAuthResponse(r.Context(), w, user.ID, false)
}

// Login handles POST /auth/login with body { "username", "password" }.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userSvc.Authenticate(r.Context(), username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "auth error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeAuthResponse(r.Context(), w, user.ID, false)
}

// DemoLogin handles POST /auth/demo and issues a read-only demo session token.
func (h *Handler) DemoLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.userSvc.Authenticate(r.Context(), h.demoUsername, h.demoPassword)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			http.Error(w, "demo account unavailable", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "auth error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeAuthResponse(r.Context(), w, user.ID, true)
}

func (h *Handler) writeAuthResponse(ctx context.Context, w http.ResponseWriter, userID string, isDemo bool) {
	profiles, err := h.profileSvc.ListByUserID(ctx, userID)
	if err != nil {
		http.Error(w, "profiles error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseProfiles := make([]authProfileResponse, 0, len(profiles))
	for _, p := range profiles {
		responseProfiles = append(responseProfiles, authProfileResponse{
			ID:     p.ID,
			DjName: p.DjName,
			Bio:    p.Bio,
		})
	}

	token, err := GenerateToken(userID, h.jwtSecret, isDemo)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{
		Token:    token,
		UserID:   userID,
		Profiles: responseProfiles,
	})
}
