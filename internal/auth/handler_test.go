package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/djbook/backend/internal/auth"
	"github.com/djbook/backend/internal/repository"
	"github.com/djbook/backend/internal/service"
	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "test-jwt-secret"

type authResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"userId"`
	Profiles []struct {
		ID     string  `json:"id"`
		DjName string  `json:"djName"`
		Bio    *string `json:"bio,omitempty"`
	} `json:"profiles"`
}

func newAuthTestHandler() (*auth.Handler, *repository.MemoryUserStore) {
	userStore := repository.NewMemoryUserStore()
	profileStore := repository.NewMemoryProfileStore()
	userSvc := service.NewUserService(userStore)
	profileSvc := service.NewProfileService(profileStore)
	return auth.NewHandler(userSvc, profileSvc, testJWTSecret), userStore
}

func postJSON(t *testing.T, handler http.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func decodeAuthResponse(t *testing.T, rec *httptest.ResponseRecorder) authResponse {
	t.Helper()
	var resp authResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v body=%q", err, rec.Body.String())
	}
	return resp
}

func TestRegister_Success(t *testing.T) {
	h, _ := newAuthTestHandler()
	rec := postJSON(t, h.Register, "/auth/register", map[string]string{
		"username": "newdj",
		"password": "secret12",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}

	resp := decodeAuthResponse(t, rec)
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.UserID == "" {
		t.Fatal("expected non-empty userId")
	}
	if len(resp.Profiles) != 1 || resp.Profiles[0].DjName != "newdj" {
		t.Fatalf("unexpected profiles: %+v", resp.Profiles)
	}
}

func TestRegister_UsernameTaken(t *testing.T) {
	h, _ := newAuthTestHandler()
	body := map[string]string{"username": "taken", "password": "secret12"}

	first := postJSON(t, h.Register, "/auth/register", body)
	if first.Code != http.StatusOK {
		t.Fatalf("first register failed: %d %q", first.Code, first.Body.String())
	}

	second := postJSON(t, h.Register, "/auth/register", body)
	if second.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", second.Code, http.StatusConflict)
	}
	if !strings.Contains(second.Body.String(), "username already taken") {
		t.Fatalf("unexpected body: %q", second.Body.String())
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	h, _ := newAuthTestHandler()
	rec := postJSON(t, h.Register, "/auth/register", map[string]string{
		"username": "shortpw",
		"password": "12345",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "password must be at least 6 characters") {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestLogin_Success(t *testing.T) {
	h, _ := newAuthTestHandler()
	creds := map[string]string{"username": "logindj", "password": "secret12"}

	reg := postJSON(t, h.Register, "/auth/register", creds)
	if reg.Code != http.StatusOK {
		t.Fatalf("register failed: %d %q", reg.Code, reg.Body.String())
	}

	rec := postJSON(t, h.Login, "/auth/login", creds)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}

	resp := decodeAuthResponse(t, rec)
	if resp.Token == "" {
		t.Fatal("expected token")
	}
	if len(resp.Profiles) != 1 {
		t.Fatalf("expected profiles array, got %+v", resp.Profiles)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	h, _ := newAuthTestHandler()
	creds := map[string]string{"username": "wrongpw", "password": "secret12"}

	reg := postJSON(t, h.Register, "/auth/register", creds)
	if reg.Code != http.StatusOK {
		t.Fatalf("register failed: %d %q", reg.Code, reg.Body.String())
	}

	rec := postJSON(t, h.Login, "/auth/login", map[string]string{
		"username": "wrongpw",
		"password": "badpass9",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), "invalid username or password") {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestRegister_EmptyUsername(t *testing.T) {
	h, _ := newAuthTestHandler()
	rec := postJSON(t, h.Register, "/auth/register", map[string]string{
		"username": "",
		"password": "secret12",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "username must be at least 3 characters") {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestRegister_PasswordHashing(t *testing.T) {
	h, userStore := newAuthTestHandler()
	password := "secret12"

	rec := postJSON(t, h.Register, "/auth/register", map[string]string{
		"username": "hashdj",
		"password": password,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("register failed: %d %q", rec.Code, rec.Body.String())
	}

	user, err := userStore.GetByUsername(context.Background(), "hashdj")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.PasswordHash == nil {
		t.Fatal("expected password hash to be stored")
	}
	if *user.PasswordHash == password {
		t.Fatal("password must not be stored as plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		t.Fatalf("bcrypt compare failed: %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	h, _ := newAuthTestHandler()
	rec := postJSON(t, h.Login, "/auth/login", map[string]string{
		"username": "ghost",
		"password": "secret12",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), "invalid username or password") {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestJWT_TokenExpiration(t *testing.T) {
	userID := "user-exp-test"
	before := time.Now()

	token, err := auth.GenerateToken(userID, testJWTSecret)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := auth.ValidateToken(token, testJWTSecret)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected exp claim")
	}

	expectedExpiry := before.Add(30 * 24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	if actualExpiry.Before(expectedExpiry.Add(-time.Minute)) || actualExpiry.After(expectedExpiry.Add(time.Minute)) {
		t.Fatalf("exp claim = %v, want ~%v", actualExpiry, expectedExpiry)
	}
}

func TestJWT_TokenContainsUserId(t *testing.T) {
	userID := "user-id-claim-test"

	token, err := auth.GenerateToken(userID, testJWTSecret)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := auth.ValidateToken(token, testJWTSecret)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("userId claim = %q, want %q", claims.UserID, userID)
	}
}
