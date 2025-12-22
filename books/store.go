package books

import (
	"context"
	"errors"
	"sync"
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
