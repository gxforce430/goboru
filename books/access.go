package books

import "context"

type MemoryAudioAccess struct {
	FullURLs    map[string]string // chapterID → full URL
	PreviewURLs map[string]string // chapterID → preview URL
}

func NewMemoryAudioAccess() *MemoryAudioAccess {
	return &MemoryAudioAccess{
		FullURLs:    make(map[string]string),
		PreviewURLs: make(map[string]string),
	}
}

func (audio *MemoryAudioAccess) PreviewURL(ctx context.Context, chapterID string) (string, error) {
	url, ok := audio.PreviewURLs[chapterID]
	if !ok {
		return "", ErrNotFound
	}

	return url, nil
}

func (audio *MemoryAudioAccess) FullURL(ctx context.Context, userID, chapterID string) (string, error) {
	if userID == "" {
		return "", ErrUserRequired
	}

	url, ok := audio.FullURLs[chapterID]
	if !ok {
		return "", ErrNotFound
	}

	return url, nil
}

// Purchase Checker
type MemoryPurchaseChecker struct {
	Owned map[string]map[string]bool
	// userID → bookID → true
}

func NewMemoryPurchaseChecker() *MemoryPurchaseChecker {
	return &MemoryPurchaseChecker{
		Owned: make(map[string]map[string]bool),
	}
}

func (checker *MemoryPurchaseChecker) HasAccess(ctx context.Context, userID, bookID string) (bool, error) {
	if userID == "" {
		return false, nil
	}

	books, ok := checker.Owned[userID]
	if !ok {
		return false, nil
	}

	return books[bookID], nil
}
