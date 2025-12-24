package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/fahmaliyi/goboruu/auth"
	"github.com/go-chi/chi/v5"
)

const sessionCookieName = "session_id"

type AuthHandler struct {
	service        *auth.Service
	googleProvider *auth.GoogleProvider
}

func NewAuthHandler(service *auth.Service, googleProvider *auth.GoogleProvider) *AuthHandler {
	return &AuthHandler{
		service:        service,
		googleProvider: googleProvider,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Get("/me", h.me)
		r.Post("/signin", h.signin)
		r.Post("/signup", h.signup)
		r.Post("/signout", h.signout)
		r.Post("/google/callback", h.googleCallback)
	})
}

func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		h.handleServiceError(w, auth.ErrUnauthorized)
		return
	}

	fmt.Println("Cookie value ", cookie.Value)

	user, session, err := h.service.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.clearSessionCookie(w)
		h.handleServiceError(w, err)
		return
	}

	h.setSessionCookie(w, session.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func (h *AuthHandler) signin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, session, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.setSessionCookie(w, session.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func (h *AuthHandler) signout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		h.handleServiceError(w, auth.ErrUnauthorized)
		return
	}

	_ = h.service.Logout(r.Context(), cookie.Value)
	h.clearSessionCookie(w)

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) signup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (h *AuthHandler) googleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "code is required"})
		return
	}

	session, err := h.service.OAuthLogin(r.Context(), h.googleProvider.Name(), code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.setSessionCookie(w, session.ID)

	writeJSON(w, http.StatusOK, map[string]string{"status": "authenticated"})
}

func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	// 1. Authentication & Token Errors (401)
	case errors.Is(err, auth.ErrUnauthorized),
		errors.Is(err, auth.ErrInvalidCredentials),
		errors.Is(err, auth.ErrTokenInvalid),
		errors.Is(err, auth.ErrTokenExpired):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})

	// 2. Conflict Errors (409)
	case errors.Is(err, auth.ErrUserAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})

	// 3. Validation Errors (400)
	case errors.Is(err, auth.ErrorEmailPasswordRequired):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})

	// 4. Resource Errors (404)
	case errors.Is(err, auth.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})

	// 5. Account Status Errors (403)
	case errors.Is(err, auth.ErrUserNotVerified):
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "please verify your email before logging in",
			"code":  "USER_NOT_VERIFIED",
		})

	// 6. Unexpected Fallback (500)
	default:
		log.Printf("[API ERROR] %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   60 * 60 * 24 * 7,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
