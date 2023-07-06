package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"sync"
)

type InMemStorage struct {
	db              map[string][]byte
	storageFilePath string
	mu              sync.Mutex
}

func NewInMemStorage(filePath string) storage.Repository {
	s := &InMemStorage{
		db:              make(map[string][]byte),
		storageFilePath: filePath,
	}

	if err := s.restoreFromFile(); err != nil {
		logger.Log.Error("cannot restore url records from file", zap.Error(err))
	}

	return s
}

func (s *InMemStorage) Put(id string, url []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[id] = url

	if err := s.writeRecordToFile(id, url); err != nil {
		logger.Log.Error("cannot write url record to file storage", zap.Error(err))
	}
}

func (s *InMemStorage) Get(id string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, ok := s.db[id]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", id)
	}

	return url, nil
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

	s.db = make(map[string][]byte)

	dec := json.NewDecoder(f)
	for dec.More() {
		var record storage.URLRecord
		if err := dec.Decode(&record); err != nil {
			return err
		}

		s.db[record.ShortURL] = []byte(record.OriginalURL)
	}

	return nil
}

func (s *InMemStorage) writeRecordToFile(id string, url []byte) error {
	if s.storageFilePath == "" {
		return nil
	}

	record := storage.URLRecord{
		UUID:        uuid.New().String(),
		ShortURL:    id,
		OriginalURL: string(url),
	}

	f, err := os.OpenFile(s.storageFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(&record); err != nil {
		return err
	}

	return nil
}
