package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Auth struct {
	JWTSecret          string
	AccessTokenTTL     int
	RefreshTokenTTL    int
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
	RefreshTokenName   string
}

type Server struct {
	Host string
	Port string
	Env  string
}

type Database struct {
	Driver string
	DSN    string
}

type Config struct {
	Auth     Auth
	Server   Server
	Database Database
}

func Load() Config {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, using system env")
	}

	return Config{
		Auth:     loadAuthConfig(),
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
	}
}

// Loaders
func loadAuthConfig() Auth {
	return Auth{
		JWTSecret:          mustGetEnv("AUTH_JWT_SECRET"),
		AccessTokenTTL:     getEnvAsInt("AUTH_ACCESS_TOKEN_TTL", 15),
		RefreshTokenTTL:    getEnvAsInt("AUTH_REFRESH_TOKEN_TTL", 43200), // 30 days
		GoogleClientID:     mustGetEnv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: mustGetEnv("GOOGLE_CLIENT_SECRET"),
		GoogleCallbackURL:  mustGetEnv("GOOGLE_CALLBACK_URL"),
		RefreshTokenName:   getEnv("REFRESH_TOKEN_NAME", "refresh_token"),
	}
}

func loadServerConfig() Server {
	return Server{
		Host: getEnv("SERVER_HOST", "0.0.0.0"),
		Port: getEnv("SERVER_PORT", "8080"),
		Env:  getEnv("APP_ENV", "development"),
	}
}

func loadDatabaseConfig() Database {
	return Database{
		Driver: getEnv("DB_DRIVER", "postgres"),
		DSN:    mustGetEnv("DB_DSN"),
	}
}

// Helpers
func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing required env: %s", key)
	}
	return value
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultVal
}
