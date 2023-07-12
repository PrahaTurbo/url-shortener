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
	db              map[string]storage.URLRecord
	storageFilePath string
	logger          *logger.Logger
	mu              sync.Mutex
}

func NewInMemStorage(filePath string, logger *logger.Logger) storage.Repository {
	s := &InMemStorage{
		db:              make(map[string]storage.URLRecord),
		storageFilePath: filePath,
		logger:          logger,
	}

	if err := s.restoreFromFile(); err != nil {
		logger.Error("cannot restore url records from file", zap.Error(err))
	}

	return s
}

func (s *InMemStorage) PutURL(_ context.Context, r storage.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[r.ShortURL] = r

	if err := s.writeRecordToFile(r); err != nil {
		return err
	}

	return nil
}

func (s *InMemStorage) PutBatchURLs(ctx context.Context, urls []storage.URLRecord) error {
	for _, r := range urls {
		if err := s.PutURL(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

func (s *InMemStorage) GetURL(_ context.Context, id string) (*storage.URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.db[id]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", id)
	}

	return &record, nil
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

	s.db = make(map[string]storage.URLRecord)

	dec := json.NewDecoder(f)
	for dec.More() {
		var r storage.URLRecord
		if err := dec.Decode(&r); err != nil {
			return err
		}

		s.db[r.ShortURL] = r
	}

	return nil
}

func (s *InMemStorage) writeRecordToFile(r storage.URLRecord) error {
	if s.storageFilePath == "" {
		return nil
	}

	s.logger.Info("write to file")

	f, err := os.OpenFile(s.storageFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	s.logger.Info("file opened")

	if err := json.NewEncoder(f).Encode(&r); err != nil {
		return err
	}

	s.logger.Info("file encoded")

	return nil
}
