package mock

import (
	"errors"
	"fmt"
)

type StorageMock struct {
	DB    map[string][]byte
	IsSQL bool
}

func (s *StorageMock) Put(id string, url []byte) {
	s.DB[id] = url
}

func (s *StorageMock) Get(id string) ([]byte, error) {
	url, ok := s.DB[id]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", id)
	}

	return url, nil
}

func (s *StorageMock) Ping() error {
	if s.IsSQL {
		return nil
	}

	return errors.New("sql database wasn't setup")
}
