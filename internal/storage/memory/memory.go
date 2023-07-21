package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"go.uber.org/zap"
	"os"
	"sync"
)

type InMemStorage struct {
	db              map[Key]*storage.URLRecord
	storageFilePath string
	logger          *logger.Logger
	mu              sync.Mutex
}

type Key struct {
	UserID   string
	ShortURL string
}

func NewInMemStorage(filePath string, logger *logger.Logger) storage.Repository {
	s := &InMemStorage{
		db:              make(map[Key]*storage.URLRecord),
		storageFilePath: filePath,
		logger:          logger,
	}

	if err := s.restoreFromFile(); err != nil {
		logger.Error("cannot restore url records from file", zap.Error(err))
	}

	return s
}

func (s *InMemStorage) PutURL(_ context.Context, r *storage.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := Key{
		UserID:   r.UserID,
		ShortURL: r.ShortURL,
	}

	s.db[k] = r

	if err := s.writeRecordToFile(r); err != nil {
		return err
	}

	return nil
}

func (s *InMemStorage) PutBatchURLs(ctx context.Context, urls []*storage.URLRecord) error {
	for _, r := range urls {
		if err := s.PutURL(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

func (s *InMemStorage) GetURL(_ context.Context, shortURL, userID string) (*storage.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := Key{
		UserID:   userID,
		ShortURL: shortURL,
	}

	record, ok := s.db[k]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", shortURL)
	}

	return record, nil
}

func (s *InMemStorage) GetURLsByUserID(_ context.Context, userID string) ([]*storage.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var records []*storage.URLRecord
	for k, v := range s.db {
		if k.UserID == userID {
			records = append(records, v)
		}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no short urls for id %s", userID)
	}

	return records, nil
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
		var r storage.URLRecord
		if err := dec.Decode(&r); err != nil {
			return err
		}

		k := Key{
			UserID:   r.UserID,
			ShortURL: r.ShortURL,
		}

		s.db[k] = &r
	}

	return nil
}

func (s *InMemStorage) writeRecordToFile(r *storage.URLRecord) error {
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
