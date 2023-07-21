package pg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"time"
)

type SQLStorage struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewSQLStorage(db *sql.DB, logger *logger.Logger) storage.Repository {
	s := &SQLStorage{
		db:     db,
		logger: logger,
	}

	return s
}

func (s *SQLStorage) PutURL(ctx context.Context, url *storage.URLRecord) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	query := `
		INSERT INTO short_urls (id, user_id, short_url, original_url)
		VALUES ($1, $2, $3, $4)`

	_, err := s.db.ExecContext(timeoutCtx, query, url.UUID, url.UserID, url.ShortURL, url.OriginalURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLStorage) PutBatchURLs(ctx context.Context, urls []*storage.URLRecord) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO short_urls (id, user_id, short_url, original_url)
		VALUES ($1, $2, $3, $4)`

	stmt, err := tx.PrepareContext(timeoutCtx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err := stmt.ExecContext(timeoutCtx, url.UUID, url.UserID, url.ShortURL, url.OriginalURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLStorage) GetURL(ctx context.Context, shortURL string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	query := `
		SELECT original_url
		FROM short_urls
		WHERE short_url = $1`

	row := s.db.QueryRowContext(timeoutCtx, query, shortURL)

	var originalURL string
	if err := row.Scan(&originalURL); err != nil {
		return "", err
	}

	if err := row.Err(); err != nil {
		return "", err
	}

	return originalURL, nil
}

func (s *SQLStorage) GetURLsByUserID(ctx context.Context, userID string) ([]*storage.URLRecord, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	query := `
		SELECT id, user_id, short_url, original_url 
		FROM short_urls
		WHERE user_id = $1`

	rows, err := s.db.QueryContext(timeoutCtx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*storage.URLRecord
	for rows.Next() {
		var r storage.URLRecord
		if err := rows.Scan(&r.UUID, &r.UserID, &r.ShortURL, &r.OriginalURL); err != nil {
			return nil, err
		}

		records = append(records, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no short urls for id %s", userID)
	}

	return records, nil
}

func (s *SQLStorage) CheckExistence(ctx context.Context, shortURL, userID string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	query := `
		SELECT id 
		FROM short_urls
		WHERE user_id = $1 AND short_url = $2`

	row := s.db.QueryRowContext(timeoutCtx, query, userID, shortURL)

	var id string
	if err := row.Scan(&id); err != nil {
		return err
	}

	if err := row.Err(); err != nil {
		return err
	}

	return nil
}

func (s *SQLStorage) Ping() error {
	return s.db.Ping()
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTable(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `
		CREATE TABLE IF NOT EXISTS short_urls (
			id UUID,
			user_id UUID,
			short_url VARCHAR PRIMARY KEY,
  			original_url VARCHAR UNIQUE,
  			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
