package main

import (
	"log"
	"time"

	"github.com/fahmaliyi/goboruu/api"
	"github.com/fahmaliyi/goboruu/auth"
	"github.com/fahmaliyi/goboruu/books"
	"github.com/fahmaliyi/goboruu/config"
)

func main() {
	cfg := config.Load()

	db, err := config.ConnectDB(cfg.Database.DATABASE_URI)
	if err != nil {
		log.Fatalf("Fatal error during database initialization: %v", err)
	}
	defer db.Close()

	hasher := auth.NewBcryptHasher(12)
	userStore := auth.NewPostgresUserStore(db)
	bookStore := books.NewPostgresBookStore(db)
	sessionStore := auth.NewPostgresSessionStore(db)

	sessionManager := auth.NewPostgresSessionManager(
		sessionStore,
		time.Duration(cfg.Auth.SessionTTL)*time.Second,
	)

	googleProvider := auth.NewGoogleProvider(
		cfg.Auth.GoogleClientID,
		cfg.Auth.GoogleClientSecret,
		cfg.Auth.GoogleCallbackURL,
	)

	authService := auth.NewService(
		userStore,
		sessionManager,
		hasher,
		googleProvider,
	)

	booksService := books.NewService(bookStore)

	api := api.NewAPI(authService, googleProvider, booksService)
	api.Run(":" + cfg.Server.Port)
}
