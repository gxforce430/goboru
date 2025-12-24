package api

import (
	"net/http"

	"github.com/fahmaliyi/goboruu/auth"
	"github.com/fahmaliyi/goboruu/books"
	"github.com/go-chi/chi/v5"
)

type BooksHandler struct {
	auth    *auth.Service
	service *books.Service
}

func NewBooksHandler(auth *auth.Service, service *books.Service) *BooksHandler {
	return &BooksHandler{
		auth:    auth,
		service: service,
	}
}

func (h *BooksHandler) RegisterRoutes(r chi.Router) {
	r.Route("/books", func(r chi.Router) {
		r.Use(AuthMiddleware(h.auth))

		r.Get("/", h.list)
		r.Get("/{bookID}", h.get)
		r.Get("/{bookID}/chapters", h.chapters)
		r.Post("/{bookID}/{chapterID}/play", h.play)
	})
}

func (h *BooksHandler) get(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookID")
	if bookID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing bookID",
		})
		return
	}

	book, err := h.service.Store.Get(r.Context(), bookID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "book not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BooksHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Store.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list books",
		})
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *BooksHandler) play(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "user required",
		})
		return
	}

	bookID := chi.URLParam(r, "bookID")
	chapterID := chi.URLParam(r, "chapterID")

	if bookID == "" || chapterID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing bookID or chapterID",
		})
		return
	}

	preview := r.URL.Query().Get("preview") == "true"

	url, err := h.service.Play(r.Context(), user.Role, user.ID, bookID, chapterID, preview)
	if err != nil {
		switch err {
		case books.ErrAccessDenied:
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error": err.Error(),
			})
		case books.ErrInvalidChapter, books.ErrNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to play audio",
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"url": url,
	})
}

func (h *BooksHandler) chapters(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookID")
	if bookID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing bookID",
		})
		return
	}

	items, err := h.service.Store.Chapters(r.Context(), bookID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "chapters not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, items)
}
