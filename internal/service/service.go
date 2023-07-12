package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/google/uuid"
)

var ErrAlready = errors.New("URL already in storage")

type Service struct {
	Storage storage.Repository
	baseURL string
}

func NewService(baseURL string, storage storage.Repository) Service {
	return Service{
		Storage: storage,
		baseURL: baseURL,
	}
}

func (s *Service) SaveURL(ctx context.Context, url string) (string, error) {
	id := s.generateID(url)

	if s.alreadyInStorage(ctx, id) {
		return s.formURL(id), ErrAlready
	}

	r := storage.URLRecord{
		UUID:        uuid.New().String(),
		ShortURL:    id,
		OriginalURL: url,
	}

	if err := s.Storage.PutURL(ctx, r); err != nil {
		return "", err
	}

	return s.formURL(id), nil
}

func (s *Service) SaveBatch(ctx context.Context, batch []models.BatchRequest) ([]models.BatchResponse, error) {
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

		if s.alreadyInStorage(ctx, id) {
			continue
		}

		var r storage.URLRecord
		r.ShortURL = s.generateID(req.OriginalURL)
		r.OriginalURL = req.OriginalURL
		r.UUID = uuid.New().String()
		records = append(records, r)
	}

	if err := s.Storage.PutBatchURLs(ctx, records); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) GetURL(ctx context.Context, id string) (string, error) {
	r, err := s.Storage.GetURL(ctx, id)
	if err != nil || r == nil {
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

func (s *Service) alreadyInStorage(ctx context.Context, id string) bool {
	if _, err := s.GetURL(ctx, id); err == nil {
		return true
	}

	return false
}
