package models

// Request represents a URL shortening request.
type Request struct {
	URL string `json:"url"`
}

// Response is the structure of a response from a URL shortening request.
type Response struct {
	Result string `json:"result"`
}

// BatchRequest represents a batch URL shortening request.
// It includes a CorrelationID for tracking and the OriginalURL that needs to be shortened.
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse is the structure of a response from a batch URL shortening request.
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLsResponse is the structure of a response containing a user's URLs.
type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLDeletionTask represents a task to delete URLs for deletion worker.
type URLDeletionTask struct {
	UserID string
	URLs   []string
}
