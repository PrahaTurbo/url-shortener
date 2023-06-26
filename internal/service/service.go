package service

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
)

type Service struct {
	URLs storage.Repository
}

func NewService(storageFilePath string) Service {
	return Service{storage.NewStorage(storageFilePath)}
}

func (s *Service) SaveURL(url []byte) string {
	// TODO Check if url has https or http prefix and add it if it doesn't
	id := s.generateID(url)
	s.URLs.Put(id, url)

	return id
}

func (s *Service) GetURL(id string) ([]byte, error) {
	return s.URLs.Get(id)
}

func (s *Service) generateID(url []byte) string {
	hasher := sha256.New()
	hasher.Write(url)
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	truncatedHash := hash[:6]

	return truncatedHash
}
