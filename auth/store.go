package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InMemoryUserStore struct {
	mu      sync.RWMutex
	byID    map[string]*User
	byEmail map[string]*User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		byID:    make(map[string]*User),
		byEmail: make(map[string]*User),
	}
}

func (s *InMemoryUserStore) Create(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byEmail[user.Email]; exists {
		return ErrUserAlreadyExists
	}

	s.byID[user.ID] = user
	s.byEmail[user.Email] = user
	return nil
}

func (s *InMemoryUserStore) FindByID(ctx context.Context, id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.byID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *InMemoryUserStore) FindByEmail(ctx context.Context, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.byEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Postgres Implementation of the store
type PostgresUserStore struct {
	db *pgxpool.Pool
}

func NewPostgresUserStore(db *pgxpool.Pool) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) Create(ctx context.Context, user *User) error {
	if user.Email == "" || (user.Provider == "local" && user.Password == "") {
		return ErrorEmailPasswordRequired
	}

	query := `
		INSERT INTO users (id, email, name, password, verified, provider, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.Exec(ctx, query,
		user.ID, user.Email, user.Name, user.Password,
		user.Verified, user.Provider, time.Now(),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (s *PostgresUserStore) FindByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, name, password, verified, provider, created_at FROM users WHERE id = $1`
	return s.scanUser(s.db.QueryRow(ctx, query, id))
}

func (s *PostgresUserStore) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, name, password, verified, provider, created_at FROM users WHERE email = $1`
	return s.scanUser(s.db.QueryRow(ctx, query, email))
}

// Internal helper to reduce code duplication
func (s *PostgresUserStore) scanUser(row pgx.Row) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.Verified, &user.Provider, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
