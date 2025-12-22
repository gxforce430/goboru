package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSessionStore struct {
	db *pgxpool.Pool
}

func NewPostgresSessionStore(db *pgxpool.Pool) *PostgresSessionStore {
	return &PostgresSessionStore{db: db}
}

func (s *PostgresSessionStore) Create(ctx context.Context, session *Session) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO sessions (id, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`,
		session.ID,
		session.UserID,
		session.CreatedAt.UTC(),
		session.ExpiresAt.UTC(),
	)
	return err
}

func (s *PostgresSessionStore) FindByID(ctx context.Context, sessionID string) (*Session, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, user_id, created_at, expires_at
		FROM sessions
		WHERE id = $1
	`, sessionID)

	var session Session
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *PostgresSessionStore) Delete(ctx context.Context, sessionID string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM sessions WHERE id = $1
	`, sessionID)
	return err
}

func (s *PostgresSessionStore) DeleteByUser(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM sessions WHERE user_id = $1
	`, userID)
	return err
}

// PostgresSessionManager implements SessionManager for Postgres-backed sessions
type PostgresSessionManager struct {
	store SessionStore
	ttl   time.Duration
}

func NewPostgresSessionManager(store SessionStore, ttl time.Duration) *PostgresSessionManager {
	return &PostgresSessionManager{
		store: store,
		ttl:   ttl,
	}
}

func (m *PostgresSessionManager) Create(ctx context.Context, userID string) (*Session, error) {
	now := time.Now().UTC()

	session := &Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(m.ttl),
	}

	if err := m.store.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (m *PostgresSessionManager) Validate(ctx context.Context, sessionID string) (*Session, error) {
	session, err := m.store.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(session.ExpiresAt.UTC()) {
		_ = m.store.Delete(ctx, sessionID)
		return nil, ErrUnauthorized
	}

	return session, nil
}

func (m *PostgresSessionManager) Rotate(ctx context.Context, oldSessionID string) (*Session, error) {
	session, err := m.Validate(ctx, oldSessionID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if err := m.store.Delete(ctx, oldSessionID); err != nil {
		return nil, err
	}

	return m.Create(ctx, session.UserID)
}

func (m *PostgresSessionManager) Destroy(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}
