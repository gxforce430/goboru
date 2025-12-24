package books

import (
	"context"
	"errors"

	"github.com/fahmaliyi/goboruu/auth"
)

var (
	ErrNotFound       = errors.New("book not found")
	ErrUserRequired   = errors.New("user required")
	ErrAccessDenied   = errors.New("full access requires purchase")
	ErrInvalidChapter = errors.New("chapter does not belong to book")
)

type Audiobook struct {
	ID          string
	Title       string
	Author      string
	Language    string
	Description string
	PriceETB    int64
	IsFree      bool
}

type Chapter struct {
	ID          string
	BookID      string
	Title       string
	Order       int
	DurationSec int
}

type Store interface {
	List(ctx context.Context) ([]Audiobook, error)
	Get(ctx context.Context, bookID string) (*Audiobook, error)
	Chapters(ctx context.Context, bookID string) ([]Chapter, error)
}

type AudioAccess interface {
	PreviewURL(ctx context.Context, chapterID string) (string, error)
	FullURL(ctx context.Context, userID, chapterID string) (string, error)
}

type PurchaseChecker interface {
	HasAccess(ctx context.Context, userID, bookID string) (bool, error)
}

// Player service
type Service struct {
	Store    Store
	Audio    AudioAccess
	Purchase PurchaseChecker
}

func NewService(store Store, audio AudioAccess, purchase PurchaseChecker) *Service {
	return &Service{
		Store:    store,
		Audio:    audio,
		Purchase: purchase,
	}
}

func (p *Service) Play(ctx context.Context, role auth.UserRole, userID, bookID, chapterID string, preview bool) (string, error) {
	chapters, err := p.Store.Chapters(ctx, bookID)
	if err != nil {
		return "", err
	}

	valid := false
	for _, ch := range chapters {
		if ch.ID == chapterID {
			valid = true
			break
		}
	}

	if !valid {
		return "", ErrInvalidChapter
	}

	if preview {
		return p.Audio.PreviewURL(ctx, chapterID)
	}

	// Admin orverride
	if role == auth.RoleAdmin {
		return p.Audio.FullURL(ctx, userID, chapterID)
	}

	book, err := p.Store.Get(ctx, bookID)
	if err != nil {
		return "", err
	}

	// Free book override
	if book.IsFree {
		return p.Audio.FullURL(ctx, userID, chapterID)
	}

	ok, err := p.Purchase.HasAccess(ctx, userID, bookID)
	if err != nil || !ok {
		return "", ErrAccessDenied
	}

	return p.Audio.FullURL(ctx, userID, chapterID)
}
