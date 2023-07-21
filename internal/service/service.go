package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/config"
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

func (s *Service) SaveURL(ctx context.Context, originalURL string) (string, error) {
	shortURL := s.generateShortURL(originalURL)

	userID, err := s.extractUserIDFromCtx(ctx)
	if err != nil {
		return "", err
	}

	if s.alreadyInStorage(ctx, shortURL, userID) {
		return s.formURL(shortURL), ErrAlready
	}

	r := &storage.URLRecord{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	if err = s.Storage.PutURL(ctx, r); err != nil {
		return "", err
	}

	return s.formURL(shortURL), nil
}

func (s *Service) SaveBatch(ctx context.Context, batch []models.BatchRequest) ([]models.BatchResponse, error) {
	records := make([]*storage.URLRecord, 0, len(batch))
	response := make([]models.BatchResponse, 0, len(batch))

	userID, err := s.extractUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	for _, req := range batch {
		if req.OriginalURL == "" {
			return nil, errors.New("no url in original_url field")
		}

		shortURL := s.generateShortURL(req.OriginalURL)

		var res models.BatchResponse
		res.CorrelationID = req.CorrelationID
		res.ShortURL = s.formURL(shortURL)
		response = append(response, res)

		if s.alreadyInStorage(ctx, shortURL, userID) {
			continue
		}

		r := &storage.URLRecord{
			UUID:        uuid.New().String(),
			ShortURL:    shortURL,
			OriginalURL: req.OriginalURL,
			UserID:      userID,
		}

		records = append(records, r)
	}

	if err := s.Storage.PutBatchURLs(ctx, records); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) GetURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, err := s.Storage.GetURL(ctx, shortURL)
	if err != nil || originalURL == "" {
		return "", err
	}

	return originalURL, nil
}

func (s *Service) GetURLsByUserID(ctx context.Context) ([]models.UserURLsResponse, error) {
	userID, err := s.extractUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	records, err := s.Storage.GetURLsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var response []models.UserURLsResponse

	for _, record := range records {
		r := models.UserURLsResponse{
			ShortURL:    s.formURL(record.ShortURL),
			OriginalURL: record.OriginalURL,
		}

		response = append(response, r)
	}

	return response, nil
}

func (s *Service) PingDB() error {
	return s.Storage.Ping()
}

func (s *Service) generateShortURL(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	truncatedHash := hash[:6]

	return truncatedHash
}

func (s *Service) formURL(shortURL string) string {
	return s.baseURL + "/" + shortURL
}

func (s *Service) alreadyInStorage(ctx context.Context, shortURL, userID string) bool {
	if err := s.Storage.CheckExistence(ctx, shortURL, userID); err == nil {
		return true
	}

	return false
}

func (s *Service) extractUserIDFromCtx(ctx context.Context) (string, error) {
	userIDVal := ctx.Value(config.UserIDKey)
	userID, ok := userIDVal.(string)
	if !ok {
		return "", fmt.Errorf("cannot extract userID from context")
	}

	return userID, nil
}
