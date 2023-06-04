package service

import (
	"math/rand"

	"github.com/PrahaTurbo/url-shortener/internal/storage"
)

type Service struct {
	DB storage.Storage
}

func (s *Service) SaveURL(url []byte) string {
	id := s.generateID(6)
	s.DB.Put(id, url)

	return id
}

func (s *Service) GetURL(id string) ([]byte, error) {
	return s.DB.Get(id)
}

func (s *Service) generateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	id := make([]byte, length)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}
