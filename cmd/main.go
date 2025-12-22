package main

import (
	"time"

	"github.com/fahmaliyi/goboruu/api"
	"github.com/fahmaliyi/goboruu/auth"
	"github.com/fahmaliyi/goboruu/books"
	"github.com/fahmaliyi/goboruu/config"
)

func main() {
	cfg := config.Load()

	hasher := auth.NewBcryptHasher(12)
	bookStore := books.NewInMemoryStore()
	userStore := auth.NewInMemoryUserStore()

	tokenManager := auth.NewJWTTokenManager(
		cfg.Auth.JWTSecret,
		time.Duration(cfg.Auth.AccessTokenTTL)*time.Second,
		time.Duration(cfg.Auth.RefreshTokenTTL)*time.Second,
	)

	googleProvider := auth.NewGoogleProvider(
		cfg.Auth.GoogleClientID,
		cfg.Auth.GoogleClientSecret,
		cfg.Auth.GoogleCallbackURL,
	)

	authService := auth.NewService(
		userStore,
		tokenManager,
		hasher,
		googleProvider,
	)

	booksService := books.NewService(bookStore)

	api := api.NewAPI(authService, googleProvider, booksService)
	api.Run(":" + cfg.Server.Port)
}
