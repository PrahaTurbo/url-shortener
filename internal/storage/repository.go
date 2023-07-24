package storage

import (
	"context"
)

type Repository interface {
	PutURL(ctx context.Context, url *URLRecord) error
	PutBatchURLs(ctx context.Context, urls []*URLRecord) error
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]*URLRecord, error)
	CheckExistence(ctx context.Context, shortURL, userID string) error
	DeleteURLBatch(urls []string, user string)
	Ping() error
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted,omitempty"`
}
