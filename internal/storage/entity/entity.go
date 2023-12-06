package entity

// URLRecord represents a URL stored in the database.
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted,omitempty"`
}

// Stats represents statistical data about the URLs and Users.
type Stats struct {
	URLs  int
	Users int
}
