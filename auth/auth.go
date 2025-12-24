package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

var (
	ErrUnauthorized            = errors.New("unauthorized")
	ErrTokenInvalid            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token expired")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUserNotVerified         = errors.New("user email is not verified")
	ErrorEmailPasswordRequired = errors.New("email and password are required")
)

type User struct {
	ID        string
	Email     string
	Name      string
	Password  string // empty if OAuth user
	Verified  bool
	Role      UserRole
	Provider  string // "local", "google", "github"
	CreatedAt time.Time
}

type OAuthUser struct {
	Email string
	Name  string
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type SessionStore interface {
	Create(ctx context.Context, session *Session) error
	Delete(ctx context.Context, sessionID string) error
	DeleteByUser(ctx context.Context, userID string) error
	FindByID(ctx context.Context, sessionID string) (*Session, error)
}

type SessionManager interface {
	Destroy(ctx context.Context, sessionID string) error
	Create(ctx context.Context, userID string) (*Session, error)
	Rotate(ctx context.Context, oldSessionID string) (*Session, error)
	Validate(ctx context.Context, sessionID string) (*Session, error)
}

type UserStore interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, plain string) error
}

type OAuthProvider interface {
	Name() string
	Authenticate(ctx context.Context, code string) (*OAuthUser, error)
}

// Services
type Service struct {
	users    UserStore
	hasher   PasswordHasher
	sessions SessionManager
	oauth    map[string]OAuthProvider
}

func NewService(users UserStore, sessions SessionManager, hasher PasswordHasher, oauthProviders ...OAuthProvider) *Service {
	providers := make(map[string]OAuthProvider)
	for _, p := range oauthProviders {
		providers[p.Name()] = p
	}

	return &Service{
		users:    users,
		sessions: sessions,
		hasher:   hasher,
		oauth:    providers,
	}
}

// Use cases
func (s *Service) Register(ctx context.Context, email, name, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, ErrorEmailPasswordRequired
	}

	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:        uuid.NewString(),
		Email:     email,
		Name:      name,
		Password:  hashed,
		Verified:  true,
		Provider:  "local",
		CreatedAt: time.Now(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, *Session, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if err := s.hasher.Compare(user.Password, password); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.Verified {
		return nil, nil, ErrUserNotVerified
	}

	session, err := s.sessions.Create(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, session, nil
}

func (s *Service) Refresh(ctx context.Context, sessionID string) (*User, *Session, error) {
	session, err := s.sessions.Validate(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("session validation failed: %w", err)
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	return user, session, nil
}

func (s *Service) Authenticate(ctx context.Context, sessionID string) (*User, error) {
	session, err := s.sessions.Validate(ctx, sessionID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	return s.users.FindByID(ctx, session.UserID)
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.sessions.Destroy(ctx, sessionID)
}

func (s *Service) OAuthLogin(ctx context.Context, providerName, code string) (*Session, error) {
	provider, ok := s.oauth[providerName]
	if !ok {
		return nil, ErrUnauthorized
	}

	oauthUser, err := provider.Authenticate(ctx, code)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByEmail(ctx, oauthUser.Email)
	if err != nil {
		user = &User{
			ID:        uuid.NewString(),
			Email:     oauthUser.Email,
			Name:      oauthUser.Name,
			Verified:  true,
			Provider:  providerName,
			CreatedAt: time.Now(),
		}

		if err := s.users.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	return s.sessions.Create(ctx, user.ID)
}

func (s *Service) Validate(ctx context.Context, sessionID string) (*User, *Session, error) {
	if sessionID == "" {
		return nil, nil, ErrUnauthorized
	}

	session, err := s.sessions.Validate(ctx, sessionID)
	if err != nil {
		return nil, nil, ErrUnauthorized
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	if !user.Verified {
		return nil, nil, ErrUnauthorized
	}

	return user, session, nil
}
