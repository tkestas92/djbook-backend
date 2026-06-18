package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	IsDemoKey contextKey = "isDemo"
)

// Claims represents the JWT payload.
type Claims struct {
	UserID string `json:"userId"`
	IsDemo bool   `json:"isDemo"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given user ID.
func GenerateToken(userID, secret string, isDemo bool) (string, error) {
	claims := Claims{
		UserID: userID,
		IsDemo: isDemo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT string.
func ValidateToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// JWTMiddleware validates the Bearer token and injects the user ID into the request context.
// It also checks Redis for token invalidation.
func JWTMiddleware(secret string, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Allow unauthenticated requests to pass through (resolvers will check)
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]

			// Check Redis invalidation list
			if rdb != nil {
				key := "invalid_token:" + tokenStr
				exists, err := rdb.Exists(r.Context(), key).Result()
				if err == nil && exists > 0 {
					http.Error(w, "token has been revoked", http.StatusUnauthorized)
					return
				}
			}

			claims, err := ValidateToken(tokenStr, secret)
			if err != nil {
				http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, IsDemoKey, claims.IsDemo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the context. Returns empty string if not authenticated.
func GetUserID(ctx context.Context) string {
	uid, _ := ctx.Value(UserIDKey).(string)
	return uid
}

// IsDemoSession returns true when the JWT was issued for the read-only demo flow.
func IsDemoSession(ctx context.Context) bool {
	isDemo, _ := ctx.Value(IsDemoKey).(bool)
	return isDemo
}

// RequireUserID returns an error if the user is not authenticated.
func RequireUserID(ctx context.Context) (string, error) {
	uid := GetUserID(ctx)
	if uid == "" {
		return "", fmt.Errorf("authentication required")
	}
	return uid, nil
}

// InvalidateToken adds a token to the Redis deny list.
func InvalidateToken(ctx context.Context, rdb *redis.Client, tokenStr string, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	return rdb.Set(ctx, "invalid_token:"+tokenStr, "1", ttl).Err()
}
