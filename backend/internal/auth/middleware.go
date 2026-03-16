package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const contextUserKey contextKey = "user"

// ContextUser holds the authenticated user's claims extracted from the JWT.
type ContextUser struct {
	UserID string
	Email  string
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RequireAuth is an HTTP middleware that validates a Bearer JWT in the
// Authorization header and stores the claims in the request context.
func RequireAuth(svc *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				jsonError(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				jsonError(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := svc.ValidateAccessToken(parts[1])
			if err != nil {
				jsonError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			user := &ContextUser{
				UserID: claims.UserID,
				Email:  claims.Email,
			}
			ctx := context.WithValue(r.Context(), contextUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser retrieves the authenticated user from the request context.
// Returns nil if no user is stored (unauthenticated request).
func GetUser(r *http.Request) *ContextUser {
	user, _ := r.Context().Value(contextUserKey).(*ContextUser)
	return user
}
