package books

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InMemoryStore struct {
	mu    sync.RWMutex
	books map[string]*Book
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		books: make(map[string]*Book),
	}
}

func (s *InMemoryStore) Create(ctx context.Context, book *Book) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.books[book.ID]; exists {
		return errors.New("id collision")
	}

	for _, existing := range s.books {
		if existing.Slug == book.Slug {
			return errors.New("slug already taken")
		}
	}

	s.books[book.ID] = book
	return nil
}

func (s *InMemoryStore) GetByID(ctx context.Context, id string) (*Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, ok := s.books[id]
	if !ok {
		return nil, ErrBookNotFound
	}

	return book, nil
}

func (s *InMemoryStore) List(ctx context.Context) ([]*Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*Book, 0, len(s.books))
	for _, book := range s.books {
		list = append(list, book)
	}

	return list, nil
}

func (s *InMemoryStore) GetBySlug(ctx context.Context, slug string) (*Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, book := range s.books {
		if book.Slug == slug {
			return book, nil
		}
	}
	return nil, ErrBookNotFound
}

// Postgres Implementation
type PostgresBookStore struct {
	db *pgxpool.Pool
}

func NewPostgresBookStore(db *pgxpool.Pool) *PostgresBookStore {
	return &PostgresBookStore{db: db}
}

func (s *PostgresBookStore) Create(ctx context.Context, b *Book) error {
	query := `
		INSERT INTO books (id, slug, title, summary, author_id, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.Exec(ctx, query, b.ID, b.Slug, b.Title, b.Summary, b.AuthorID, b.Description, b.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create book: %w", err)
	}
	return nil
}

func (s *PostgresBookStore) List(ctx context.Context) ([]*Book, error) {
	query := `SELECT id, slug, title, summary, author_id, description, created_at FROM books ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*Book
	for rows.Next() {
		var b Book
		err := rows.Scan(&b.ID, &b.Slug, &b.Title, &b.Summary, &b.AuthorID, &b.Description, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, &b)
	}
	return books, nil
}

func (s *PostgresBookStore) GetByID(ctx context.Context, id string) (*Book, error) {
	query := `SELECT id, slug, title, summary, author_id, description, created_at FROM books WHERE id = $1`
	return s.scanRow(s.db.QueryRow(ctx, query, id))
}

func (s *PostgresBookStore) GetBySlug(ctx context.Context, slug string) (*Book, error) {
	query := `SELECT id, slug, title, summary, author_id, description, created_at FROM books WHERE slug = $1`
	return s.scanRow(s.db.QueryRow(ctx, query, slug))
}

// Helper to avoid duplicate scanning logic
func (s *PostgresBookStore) scanRow(row pgx.Row) (*Book, error) {
	var b Book
	err := row.Scan(&b.ID, &b.Slug, &b.Title, &b.Summary, &b.AuthorID, &b.Description, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	return &b, nil
}
