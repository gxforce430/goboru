package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
	Provider  string // "local", "google", "github"
	CreatedAt time.Time
}

type OAuthUser struct {
	Email string
	Name  string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Contracts
type UserStore interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, plain string) error
}

type TokenManager interface {
	Generate(userID string) (*TokenPair, error)
	VerifyAccess(token string) (userID string, err error)
	VerifyRefresh(token string) (userID string, err error)
}

type OAuthProvider interface {
	Name() string
	Authenticate(ctx context.Context, code string) (*OAuthUser, error)
}

// Services
type Service struct {
	users  UserStore
	tokens TokenManager
	hasher PasswordHasher
	oauth  map[string]OAuthProvider
}

func NewService(users UserStore, tokens TokenManager, hasher PasswordHasher, oauthProviders ...OAuthProvider) *Service {
	providers := make(map[string]OAuthProvider)
	for _, p := range oauthProviders {
		providers[p.Name()] = p
	}

	return &Service{
		users:  users,
		tokens: tokens,
		hasher: hasher,
		oauth:  providers,
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

func (s *Service) Login(ctx context.Context, email, password string) (*User, *TokenPair, error) {
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

	tokens, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*User, *TokenPair, error) {
	userID, err := s.tokens.VerifyRefresh(refreshToken)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	tokens, err := s.tokens.Generate(userID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *Service) Authenticate(ctx context.Context, accessToken string) (*User, error) {
	userID, err := s.tokens.VerifyAccess(accessToken)
	if err != nil {
		return nil, err
	}

	return s.users.FindByID(context.Background(), userID)
}

func (s *Service) OAuthLogin(ctx context.Context, providerName, code string) (*TokenPair, error) {
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
		user := &User{
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

	return s.tokens.Generate(user.ID)
}
