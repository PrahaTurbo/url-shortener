package storage

import (
	"context"

	"github.com/PrahaTurbo/url-shortener/internal/storage/entity"
)

// Repository is an interface that defines operations to interact with the storage system.
type Repository interface {
	SaveURL(ctx context.Context, url entity.URLRecord) error
	SaveURLBatch(ctx context.Context, urls []*entity.URLRecord) error
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]entity.URLRecord, error)
	CheckExistence(ctx context.Context, shortURL, userID string) error
	DeleteURLBatch(urls []string, user string) error
	Ping() error
}
