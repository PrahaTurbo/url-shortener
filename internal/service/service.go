package service

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/PrahaTurbo/url-shortener/internal/storage"
)

type Service struct {
	DB storage.Storage
}

func (s *Service) SaveURL(url []byte) string {
	id := s.generateID(url)
	s.DB.Put(id, url)

	return id
}

func (s *Service) GetURL(id string) ([]byte, error) {
	return s.DB.Get(id)
}

func (s *Service) generateID(url []byte) string {
	hasher := sha256.New()
	hasher.Write(url)
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	truncatedHash := hash[:6]

	return truncatedHash
}
