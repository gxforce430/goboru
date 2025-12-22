package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fahmaliyi/goboruu/auth"
)

type contextKey string

const userKey contextKey = "user"

func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid token"})
				return
			}

			token := authHeader[7:]
			user, err := authService.Authenticate(r.Context(), token)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
				return
			}

			ctx := r.Context()
			ctx = contextWithUser(ctx, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helpers
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func contextWithUser(ctx context.Context, user *auth.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func userFromContext(ctx context.Context) *auth.User {
	user, _ := ctx.Value(userKey).(*auth.User)
	return user
}
