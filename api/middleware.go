package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fahmaliyi/goboruu/auth"
)

type contextKey string

const (
	userKey    contextKey = "user"
	sessionKey contextKey = "session"
)

func AuthMiddleware(service *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil || cookie.Value == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "missing or invalid session",
				})
				return
			}

			// validate session & get user
			user, session, err := service.Validate(r.Context(), cookie.Value)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "unauthorized",
				})
				return
			}

			ctx := r.Context()
			ctx = contextWithUser(ctx, user)
			ctx = contextWithSession(ctx, session)

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

func contextWithSession(ctx context.Context, session *auth.Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func userFromContext(ctx context.Context) *auth.User {
	user, _ := ctx.Value(userKey).(*auth.User)
	return user
}

func SessionFromContext(ctx context.Context) *auth.Session {
	session, _ := ctx.Value(sessionKey).(*auth.Session)
	return session
}
