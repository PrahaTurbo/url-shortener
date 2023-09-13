package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/internal/storage/entity"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
)

type InMemStorage struct {
	urls            map[string]string
	users           map[string][]entity.URLRecord
	storageFilePath string
	logger          *logger.Logger
	mu              sync.Mutex
}

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

func (s *InMemStorage) SaveURLBatch(ctx context.Context, urls []*entity.URLRecord) error {
	for _, r := range urls {
		if err := s.SaveURL(ctx, *r); err != nil {
			return err
		}
	}

	return nil
}

func (s *InMemStorage) GetURL(_ context.Context, shortURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalURL, ok := s.urls[shortURL]
	if !ok {
		return "", fmt.Errorf("no url for id: %s", shortURL)
	}

	return originalURL, nil
}

func (s *InMemStorage) GetURLsByUserID(_ context.Context, userID string) ([]entity.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, ok := s.users[userID]
	if !ok {
		return nil, fmt.Errorf("no short urls for id %s", userID)
	}

	return records, nil
}

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
