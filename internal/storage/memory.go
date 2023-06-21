package storage

import (
	"fmt"
	"sync"
)

type Repository interface {
	Put(id string, url []byte)
	Get(id string) ([]byte, error)
}

type Storage struct {
	DB map[string][]byte
	mu sync.Mutex
}

func (s *Storage) Put(id string, url []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.DB[id] = url
}

func (s *Storage) Get(id string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, ok := s.DB[id]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", id)
	}

	return url, nil
}

func NewStorage() Repository {
	return &Storage{DB: make(map[string][]byte)}
}
