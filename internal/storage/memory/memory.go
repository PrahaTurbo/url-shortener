package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/entity"
)

// InMemStorage maintains an in-memory representation of URL shortening data.
type InMemStorage struct {
	urls            map[string]string
	users           map[string][]entity.URLRecord
	storageFilePath string
	logger          *logger.Logger
	mu              sync.Mutex
}

// NewInMemStorage initializes a new InMemStorage instance with provided inputs
// and restore previous URL shortening data from the file if it exists.
func NewInMemStorage(filePath string, logger *logger.Logger) storage.Repository {
	s := &InMemStorage{
		urls:            make(map[string]string),
		users:           make(map[string][]entity.URLRecord),
		storageFilePath: filePath,
		logger:          logger,
	}

	if err := s.restoreFromFile(); err != nil {
		logger.Error("cannot restore url records from file", zap.Error(err))
	}

	return s
}

// SaveURL stores a new URL record in InMemStorage and writes the record to
// the file if a filePath was specified during initialization.
func (s *InMemStorage) SaveURL(_ context.Context, r entity.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[r.ShortURL] = r.OriginalURL
	s.users[r.UserID] = append(s.users[r.UserID], r)

	if err := s.writeRecordToFile(r); err != nil {
		return err
	}

	return nil
}

// SaveURLBatch stores a batch of URL records in InMemStorage by iterating through the input
// and calling SaveURL function for each record.
func (s *InMemStorage) SaveURLBatch(ctx context.Context, urls []*entity.URLRecord) error {
	for _, r := range urls {
		if err := s.SaveURL(ctx, *r); err != nil {
			return err
		}
	}

	return nil
}

// GetURL retrieves the original URL from InMemStorage given its shortened version.
func (s *InMemStorage) GetURL(_ context.Context, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, ok := s.urls[shortURL]
	if !ok {
		return "", fmt.Errorf("no url for id: %s", shortURL)
	}

	return originalURL, nil
}

// GetURLsByUserID retrieves all the URL records of a specific user from InMemStorage.
func (s *InMemStorage) GetURLsByUserID(_ context.Context, userID string) ([]entity.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, ok := s.users[userID]
	if !ok {
		return nil, fmt.Errorf("no short urls for id %s", userID)
	}

	return records, nil
}

// DeleteURLBatch marks a set of URLs associated with a user as deleted in InMemStorage
// by iterating through the user’s URL records and setting DeletedFlag to true for
// matching URLs.
func (s *InMemStorage) DeleteURLBatch(urls []string, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, ok := s.users[user]
	if !ok {
		return errors.New("user not found")
	}

	for i := range records {
		for _, url := range urls {
			if url == records[i].ShortURL {
				records[i].DeletedFlag = true
			}
		}
	}

	return nil
}

// CheckExistence checks if a shortened URL associated with a user exists in InMemStorage.
func (s *InMemStorage) CheckExistence(_ context.Context, shortURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user not found")
	}

	for _, r := range records {
		if r.ShortURL == shortURL {
			return nil
		}
	}

	return fmt.Errorf("no urls for user")
}

// Ping returns an error as InMemStorage does not maintain connection to any external
// SQL database.
func (s *InMemStorage) Ping() error {
	return errors.New("no connection to sql database")
}

func (s *InMemStorage) restoreFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.storageFilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	for dec.More() {
		var r entity.URLRecord
		if err := dec.Decode(&r); err != nil {
			return err
		}

		s.urls[r.ShortURL] = r.OriginalURL
		s.users[r.UserID] = append(s.users[r.UserID], r)
	}

	return nil
}

func (s *InMemStorage) writeRecordToFile(r entity.URLRecord) error {
	if s.storageFilePath == "" {
		return nil
	}

	f, err := os.OpenFile(s.storageFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(r); err != nil {
		return err
	}

	return nil
}
