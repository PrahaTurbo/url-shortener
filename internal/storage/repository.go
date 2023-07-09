package storage

type Repository interface {
	PutURL(url URLRecord) error
	PutBatchURLs(urls []URLRecord) error
	GetURL(id string) (*URLRecord, error)
	Ping() error
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
