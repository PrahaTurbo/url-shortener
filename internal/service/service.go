package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/entity"
)

// Service is an interface for API that handle URL shortening and associated operations.
type Service interface {
	SaveURL(ctx context.Context, originalURL string) (string, error)
	SaveBatch(ctx context.Context, batch []models.BatchRequest) ([]models.BatchResponse, error)
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetURLsByUserID(ctx context.Context) ([]models.UserURLsResponse, error)
	DeleteURLs(ctx context.Context, urls []string) error
	PingDB() error
}

type service struct {
	Storage   storage.Repository
	logger    *logger.Logger
	baseURL   string
	delChan   chan models.URLDeletionTask
	semaphore *semaphore
}

// NewService creates a new instance of the URL service with specified configurations.
func NewService(baseURL string, storage storage.Repository, logger *logger.Logger) Service {
	s := &service{
		Storage:   storage,
		logger:    logger,
		baseURL:   baseURL,
		delChan:   make(chan models.URLDeletionTask, 10),
		semaphore: newSemaphore(5),
	}

	go s.startURLDeletionWorker(time.Second*10, 100)

	return s
}

// SaveURL saves an original URL, provides a shortened version, and returns it.
func (s *service) SaveURL(ctx context.Context, originalURL string) (string, error) {
	shortURL := generateShortURL(originalURL)

	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		return "", err
	}

	if s.alreadyInStorage(ctx, shortURL, userID) {
		return formURL(s.baseURL, shortURL), ErrAlready
	}

	r := entity.URLRecord{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	if err = s.Storage.SaveURL(ctx, r); err != nil {
		return "", err
	}

	return formURL(s.baseURL, shortURL), nil
}

// SaveBatch handles the saving of multiple URLs at once, returning an array of responses.
func (s *service) SaveBatch(ctx context.Context, batch []models.BatchRequest) ([]models.BatchResponse, error) {
	records := make([]*entity.URLRecord, 0, len(batch))
	response := make([]models.BatchResponse, 0, len(batch))

	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	for _, req := range batch {
		if req.OriginalURL == "" {
			return nil, ErrNoOriginalURL
		}

		shortURL := generateShortURL(req.OriginalURL)

		var res models.BatchResponse
		res.CorrelationID = req.CorrelationID
		res.ShortURL = formURL(s.baseURL, shortURL)
		response = append(response, res)

		if s.alreadyInStorage(ctx, shortURL, userID) {
			continue
		}

		r := &entity.URLRecord{
			UUID:        uuid.New().String(),
			ShortURL:    shortURL,
			OriginalURL: req.OriginalURL,
			UserID:      userID,
		}

		records = append(records, r)
	}

	if err := s.Storage.SaveURLBatch(ctx, records); err != nil {
		return nil, err
	}

	return response, nil
}

// GetURL retrieves the original URL given the shortened one.
// If the shortened URL is not registered in the service, the error will be returned.
func (s *service) GetURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, err := s.Storage.GetURL(ctx, shortURL)
	if err != nil || originalURL == "" {
		return "", err
	}

	return originalURL, nil
}

// GetURLsByUserID retrieves all the URLs associated with a specific user.
func (s *service) GetURLsByUserID(ctx context.Context) ([]models.UserURLsResponse, error) {
	userID, err := extractUserIDFromCtx(ctx)
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
			ShortURL:    formURL(s.baseURL, record.ShortURL),
			OriginalURL: record.OriginalURL,
		}

		response = append(response, r)
	}

	return response, nil
}

// DeleteURLs initiates the process of deleting a set of URLs, passes models.URLDeletionTask to
// deletion channel.
func (s *service) DeleteURLs(ctx context.Context, urls []string) error {
	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		return err
	}

	task := models.URLDeletionTask{
		UserID: userID,
		URLs:   urls,
	}

	s.delChan <- task

	return nil
}

// PingDB checks the current status of the database.
func (s *service) PingDB() error {
	return s.Storage.Ping()
}

func (s *service) startURLDeletionWorker(interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)

	var tasks []models.URLDeletionTask

	for {
		select {
		case task := <-s.delChan:
			tasks = append(tasks, task)

			if len(tasks) >= batchSize {
				s.handleDeletion(tasks)
				tasks = nil
			}
		case <-ticker.C:
			if len(tasks) > 0 {
				s.handleDeletion(tasks)
				tasks = nil
			}
		}
	}
}

func (s *service) handleDeletion(tasks []models.URLDeletionTask) {
	for _, task := range tasks {
		s.semaphore.acquire()

		go func(urls []string, user string) {
			defer s.semaphore.release()

			if err := s.Storage.DeleteURLBatch(urls, user); err != nil {
				s.logger.Error("cannot delete batch urls", zap.Error(err), zap.String("user id", user))
			}

		}(task.URLs, task.UserID)
	}
}

func (s *service) alreadyInStorage(ctx context.Context, shortURL, userID string) bool {
	if err := s.Storage.CheckExistence(ctx, shortURL, userID); err == nil {
		return true
	}

	return false
}
