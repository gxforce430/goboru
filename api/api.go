package api

import (
	"log"
	"net/http"
	"time"

	"github.com/fahmaliyi/goboruu/auth"
	"github.com/fahmaliyi/goboruu/books"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type API struct {
	Router *chi.Mux
}

func NewAPI(authService *auth.Service, google *auth.GoogleProvider, bookService *books.Service) *API {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(10 * time.Second))

	authHandler := NewAuthHandler(authService, google)
	authHandler.RegisterRoutes(router)

	bookHandler := NewBookHandler(bookService, authService)
	bookHandler.RegisterRoutes(router)

	return &API{
		Router: router,
	}
}

func (a *API) Run(addr string) {
	srv := &http.Server{
		Addr:         addr,
		Handler:      a.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("API running on", addr)
	log.Fatal(srv.ListenAndServe())
}
