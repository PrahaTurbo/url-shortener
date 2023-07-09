package memory

import (
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
	mu              sync.Mutex
}

func NewInMemStorage(filePath string) storage.Repository {
	s := &InMemStorage{
		db:              make(map[string]storage.URLRecord),
		storageFilePath: filePath,
	}

	if err := s.restoreFromFile(); err != nil {
		logger.Log.Error("cannot restore url records from file", zap.Error(err))
	}

	return s
}

func (s *InMemStorage) PutURL(r storage.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[r.ShortURL] = r

	if err := s.writeRecordToFile(r); err != nil {
		return err
	}

	return nil
}

func (s *InMemStorage) PutBatchURLs(urls []storage.URLRecord) error {
	for _, r := range urls {
		if err := s.PutURL(r); err != nil {
			return err
		}
	}

	return nil
}

func (s *InMemStorage) GetURL(id string) (*storage.URLRecord, error) {
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

	logger.Log.Info("write to file")

	f, err := os.OpenFile(s.storageFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	logger.Log.Info("file opened")

	if err := json.NewEncoder(f).Encode(&r); err != nil {
		return err
	}

	logger.Log.Info("file encoded")

	return nil
}
