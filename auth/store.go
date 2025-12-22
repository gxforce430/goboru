package auth

import (
	"context"
	"sync"
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
