package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fahmaliyi/goboruu/auth"
	"github.com/fahmaliyi/goboruu/books"
	"github.com/go-chi/chi/v5"
)

type createBookRequest struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type BookHandler struct {
	service     *books.Service
	authService *auth.Service
}

func NewBookHandler(service *books.Service, authService *auth.Service) *BookHandler {
	return &BookHandler{
		service:     service,
		authService: authService,
	}
}

func (h *BookHandler) RegisterRoutes(r chi.Router) {
	r.Route("/books", func(r chi.Router) {
		r.Use(AuthMiddleware(h.authService))

		r.Get("/", h.listBooks)
		r.Post("/", h.createBook)
		r.Get("/{identifier}", h.getBook)
	})
}

func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r.Context())

	var req createBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	book, err := h.service.CreateBook(r.Context(), req.Title, req.Description, user.ID, req.Summary)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.ListBooks(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (h *BookHandler) getBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "identifier")
	book, err := h.service.GetBook(r.Context(), id)
	if err != nil {
		if errors.Is(err, books.ErrBookNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, book)
}
