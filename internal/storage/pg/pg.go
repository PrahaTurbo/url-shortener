package pg

import (
	"context"
	"database/sql"
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

func (s *SQLStorage) PutURL(ctx context.Context, url storage.URLRecord) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	query := `
		INSERT INTO short_urls (id, short_url, original_url)
		VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(timeoutCtx, query, url.UUID, url.ShortURL, url.OriginalURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLStorage) PutBatchURLs(ctx context.Context, urls []storage.URLRecord) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO short_urls (id, short_url, original_url)
		VALUES ($1, $2, $3)`

	stmt, err := tx.PrepareContext(timeoutCtx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err := stmt.ExecContext(timeoutCtx, url.UUID, url.ShortURL, url.OriginalURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLStorage) GetURL(ctx context.Context, shortURL string) (*storage.URLRecord, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	query := `
		SELECT id, short_url, original_url 
		FROM short_urls
		WHERE short_url = $1`

	row := s.db.QueryRowContext(timeoutCtx, query, shortURL)

	var url storage.URLRecord
	if err := row.Scan(&url.UUID, &url.ShortURL, &url.OriginalURL); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return &url, nil
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

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		CREATE TABLE IF NOT EXISTS short_urls (
			id UUID,
			short_url VARCHAR PRIMARY KEY,
  			original_url VARCHAR UNIQUE,
  			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`
	tx.ExecContext(ctx, query)

	tx.Commit()

	return nil
}
