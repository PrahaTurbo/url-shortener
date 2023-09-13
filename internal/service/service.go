package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
)

var ErrAlready = errors.New("URL already in storage")

type urlDeletionTask struct {
	userID string
	urls   []string
}

type Service struct {
	Storage   storage.Repository
	logger    *logger.Logger
	baseURL   string
	delChan   chan urlDeletionTask
	semaphore *semaphore
}

func NewService(baseURL string, storage storage.Repository, logger *logger.Logger) Service {
	s := Service{
		Storage:   storage,
		logger:    logger,
		baseURL:   baseURL,
		delChan:   make(chan urlDeletionTask, 10),
		semaphore: newSemaphore(5),
	}

	go s.startURLDeletionWorker(time.Second*10, 100)

	return s
}

func (s *Service) SaveURL(ctx context.Context, originalURL string) (string, error) {
	shortURL := generateShortURL(originalURL)

	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		return "", err
	}

	if s.alreadyInStorage(ctx, shortURL, userID) {
		return formURL(s.baseURL, shortURL), ErrAlready
	}

	r := storage.URLRecord{
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

func (s *Service) SaveBatch(ctx context.Context, batch []models.BatchRequest) ([]models.BatchResponse, error) {
	records := make([]*storage.URLRecord, 0, len(batch))
	response := make([]models.BatchResponse, 0, len(batch))

	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	for _, req := range batch {
		if req.OriginalURL == "" {
			return nil, errors.New("no url in original_url field")
		}

		shortURL := generateShortURL(req.OriginalURL)

		var res models.BatchResponse
		res.CorrelationID = req.CorrelationID
		res.ShortURL = formURL(s.baseURL, shortURL)
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

	if err := s.Storage.SaveURLBatch(ctx, records); err != nil {
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

func (s *Service) DeleteURLs(ctx context.Context, urls []string) error {
	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		return err
	}

	task := urlDeletionTask{
		userID: userID,
		urls:   urls,
	}

	s.delChan <- task

	return nil
}

func (s *Service) startURLDeletionWorker(interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)

	var tasks []urlDeletionTask

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

func (s *Service) handleDeletion(tasks []urlDeletionTask) {
	for _, task := range tasks {
		s.semaphore.acquire()

		go func(urls []string, user string) {
			defer s.semaphore.release()

			if err := s.Storage.DeleteURLBatch(urls, user); err != nil {
				s.logger.Error("cannot delete batch urls", zap.Error(err), zap.String("user id", user))
			}

		}(task.urls, task.userID)
	}
}

func (s *Service) PingDB() error {
	return s.Storage.Ping()
}

func (s *Service) alreadyInStorage(ctx context.Context, shortURL, userID string) bool {
	if err := s.Storage.CheckExistence(ctx, shortURL, userID); err == nil {
		return true
	}

	return false
}
