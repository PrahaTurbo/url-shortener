package storage

type Repository interface {
	Put(id string, url []byte)
	Get(id string) ([]byte, error)
	Ping() error
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
