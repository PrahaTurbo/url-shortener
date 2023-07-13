package storage

import "context"

type Repository interface {
	PutURL(ctx context.Context, url URLRecord) error
	PutBatchURLs(ctx context.Context, urls []URLRecord) error
	GetURL(ctx context.Context, id string) (*URLRecord, error)
	Ping() error
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
