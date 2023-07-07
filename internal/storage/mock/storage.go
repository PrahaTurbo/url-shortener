package mock

import (
	"errors"
	"fmt"
)

type StorageMock struct {
	DB    map[string]string
	IsSQL bool
}

func (s *StorageMock) Put(id string, url string) {
	s.DB[id] = url
}

func (s *StorageMock) Get(id string) (string, error) {
	url, ok := s.DB[id]
	if !ok {
		return "", fmt.Errorf("no url for id: %s", id)
	}

	return url, nil
}

func (s *StorageMock) Ping() error {
	if s.IsSQL {
		return nil
	}

	return errors.New("sql database wasn't setup")
}
