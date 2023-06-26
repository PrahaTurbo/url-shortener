package storage

import (
	"encoding/json"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"sync"
)

type Repository interface {
	Put(id string, url []byte)
	Get(id string) ([]byte, error)
}

type Storage struct {
	db              map[string][]byte
	storageFilePath string
	mu              sync.Mutex
}

type urlRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewStorage(filePath string) Repository {
	s := &Storage{
		db:              make(map[string][]byte),
		storageFilePath: filePath,
	}

	if err := s.restoreFromFile(); err != nil {
		logger.Log.Error("cannot restore url records from file", zap.Error(err))
	}
	
	return s
}

func (s *Storage) Put(id string, url []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[id] = url

	record := urlRecord{
		UUID:        uuid.New().String(),
		ShortURL:    id,
		OriginalURL: string(url),
	}

	if s.storageFilePath != "" {
		if err := s.writeRecordToFile(record); err != nil {
			logger.Log.Error("cannot write url record to file storage", zap.Error(err))
		}
	}
}

func (s *Storage) Get(id string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, ok := s.db[id]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", id)
	}

	return url, nil
}

func (s *Storage) restoreFromFile() error {
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
		var record urlRecord
		if err := dec.Decode(&record); err != nil {
			return err
		}

		s.db[record.ShortURL] = []byte(record.OriginalURL)
	}

	return nil
}

func (s *Storage) writeRecordToFile(record urlRecord) error {
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
