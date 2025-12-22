package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(connStr string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	if err := initSchema(ctx, pool); err != nil {
		return nil, fmt.Errorf("schema init failed: %w", err)
	}

	return pool, nil
}

func initSchema(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		password TEXT,
		verified BOOLEAN DEFAULT FALSE,
		provider TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		slug TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		summary TEXT,
		author_id TEXT NOT NULL REFERENCES users(id),
		description TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_books_slug ON books(slug);

	CREATE TABLE IF NOT EXISTS sessions (
		id         UUID PRIMARY KEY,
		user_id    UUID NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`
	_, err := pool.Exec(ctx, query)
	return err
}
