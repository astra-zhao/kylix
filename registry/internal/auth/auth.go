package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"kylix/registry/internal/db"
	"kylix/registry/internal/models"
)

type Service struct {
	store db.Store
}

func NewService(store db.Store) *Service {
	return &Service{store: store}
}

// GenerateToken creates a new API token (32 random bytes, hex-encoded).
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidateToken checks if the given token exists and returns the user.
func (s *Service) ValidateToken(token string) (*models.User, error) {
	return s.store.GetUserByToken(token)
}

// RequireAuth is a middleware that enforces API token authentication.
func (s *Service) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid Authorization format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		user, err := s.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error":"token validation failed"}`, http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Store user in request context (optional: for handlers that need user info)
		// For now, just proceed
		next(w, r)
	}
}
