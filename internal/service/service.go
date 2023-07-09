package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/google/uuid"
)

type Service struct {
	Storage storage.Repository
	baseURL string
}

func NewService(cfg config.Config, storage storage.Repository) Service {
	return Service{
		Storage: storage,
		baseURL: cfg.BaseURL,
	}
}

func (s *Service) SaveURL(url string) (string, error) {
	id := s.generateID(url)

	if s.alreadyInStorage(id) {
		return s.formURL(id), nil
	}

	r := storage.URLRecord{
		UUID:        uuid.New().String(),
		ShortURL:    id,
		OriginalURL: url,
	}

	if err := s.Storage.PutURL(r); err != nil {
		return "", err
	}

	return s.formURL(id), nil
}

func (s *Service) SaveBatch(batch []models.BatchRequest) ([]models.BatchResponse, error) {
	records := make([]storage.URLRecord, 0, len(batch))
	response := make([]models.BatchResponse, 0, len(batch))

	for _, req := range batch {
		if req.OriginalURL == "" {
			return nil, errors.New("no url in original_url field")
		}

		id := s.generateID(req.OriginalURL)

		var res models.BatchResponse
		res.CorrelationID = req.CorrelationID
		res.ShortURL = s.formURL(id)
		response = append(response, res)

		if s.alreadyInStorage(id) {
			continue
		}

		var r storage.URLRecord
		r.ShortURL = s.generateID(req.OriginalURL)
		r.OriginalURL = req.OriginalURL
		r.UUID = uuid.New().String()
		records = append(records, r)
	}

	if err := s.Storage.PutBatchURLs(records); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) GetURL(id string) (string, error) {
	r, err := s.Storage.GetURL(id)
	if err != nil {
		return "", err
	}

	return r.OriginalURL, nil
}

func (s *Service) PingDB() error {
	return s.Storage.Ping()
}

func (s *Service) generateID(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	truncatedHash := hash[:6]

	return truncatedHash
}

func (s *Service) formURL(id string) string {
	return s.baseURL + "/" + id
}

func (s *Service) alreadyInStorage(id string) bool {
	if _, err := s.GetURL(id); err == nil {
		return true
	}

	return false
}
