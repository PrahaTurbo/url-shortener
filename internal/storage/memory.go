package storage

import (
	"fmt"
)

type Repository interface {
	Put(id string, url []byte)
	Get(id string) ([]byte, error)
}

type Storage struct {
	DB map[string][]byte
}

func (s *Storage) Put(id string, url []byte) {
	s.DB[id] = url
}

func (s *Storage) Get(id string) ([]byte, error) {
	url, ok := s.DB[id]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", id)
	}

	return url, nil
}
