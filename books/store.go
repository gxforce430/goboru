package books

import "context"

type MemoryStore struct {
	Books        map[string]Audiobook
	BookChapters map[string][]Chapter
}

func NewMemoryStore() *MemoryStore {
	book := Audiobook{
		ID:          "book1",
		Title:       "Sample Afan Oromo Audiobook",
		Author:      "Fahm",
		Language:    "Afan Oromo",
		Description: "An example book for testing",
		PriceETB:    100,
		IsFree:      false,
	}

	chapters := []Chapter{
		{ID: "ch1", BookID: "book1", Title: "Chapter 1", Order: 1, DurationSec: 600},
		{ID: "ch2", BookID: "book1", Title: "Chapter 2", Order: 2, DurationSec: 700},
	}

	return &MemoryStore{
		Books: map[string]Audiobook{
			book.ID: book,
		},
		BookChapters: map[string][]Chapter{
			book.ID: chapters,
		},
	}
}
func (store *MemoryStore) List(ctx context.Context) ([]Audiobook, error) {
	var audiobook []Audiobook
	for _, book := range store.Books {
		audiobook = append(audiobook, book)
	}

	return audiobook, nil
}

func (store *MemoryStore) Get(ctx context.Context, bookID string) (*Audiobook, error) {
	book, ok := store.Books[bookID]
	if !ok {
		return nil, ErrNotFound
	}

	return &book, nil
}

func (store *MemoryStore) Chapters(ctx context.Context, bookID string) ([]Chapter, error) {
	chapter, ok := store.BookChapters[bookID]
	if !ok {
		return nil, ErrNotFound
	}

	return chapter, nil
}
