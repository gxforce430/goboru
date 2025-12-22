package books

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrBookNotFound = errors.New("book not found")

type Book struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	AuthorID    string    `json:"author_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Contracts
type Store interface {
	List(ctx context.Context) ([]*Book, error)
	Create(ctx context.Context, book *Book) error
	GetByID(ctx context.Context, id string) (*Book, error)
	GetBySlug(ctx context.Context, slug string) (*Book, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateBook(ctx context.Context, title, description, authorID, summary string) (*Book, error) {
	baseSlug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	baseSlug = filterSpecialChars(baseSlug)

	finalSlug := baseSlug

	existing, err := s.store.GetBySlug(ctx, finalSlug)
	if err == nil && existing != nil {
		suffix := uuid.NewString()[:4]
		finalSlug = fmt.Sprintf("%s-%s", baseSlug, suffix)
	}

	book := &Book{
		ID:          uuid.NewString(),
		Slug:        finalSlug,
		Title:       title,
		Description: description,
		AuthorID:    authorID,
		CreatedAt:   time.Now(),
		Summary:     summary,
	}

	if err := s.store.Create(ctx, book); err != nil {
		return nil, err
	}
	return book, nil
}

func filterSpecialChars(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, s)
}

func (s *Service) GetBook(ctx context.Context, identifier string) (*Book, error) {
	log.Printf("Searching for book with identifier: %s", identifier)

	book, err := s.store.GetByID(ctx, identifier)
	if err != nil {
		log.Printf("ID lookup failed for %s, trying slug...", identifier)
		return s.store.GetBySlug(ctx, identifier)
	}
	return book, nil
}

func (s *Service) ListBooks(ctx context.Context) ([]*Book, error) {
	return s.store.List(ctx)
}
