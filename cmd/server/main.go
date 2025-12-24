package main

import (
	"log"
	"time"

	"github.com/fahmaliyi/goboruu/auth"
	"github.com/fahmaliyi/goboruu/books"
	"github.com/fahmaliyi/goboruu/cmd/api"
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
	sessionStore := auth.NewPostgresSessionStore(db)
	bookStore := books.NewMemoryStore()
	audioAccess := books.NewMemoryAudioAccess()
	purchaseChecker := books.NewMemoryPurchaseChecker()

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

	booksService := books.NewService(
		bookStore,
		audioAccess,
		purchaseChecker,
	)

	api := api.NewAPI(authService, googleProvider, booksService)
	api.Run(":" + cfg.Server.Port)
}
