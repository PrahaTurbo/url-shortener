package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/entity"
)

// ErrURLDeleted is thrown when the URL being accessed has been deleted.
var ErrURLDeleted = errors.New("url was deleted")

// SQLStorage is a struct that implements the storage.Repository interface, using Postgresql as a storage backend.
type SQLStorage struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewSQLStorage initializes a new SQLStorage instance with provided inputs.
func NewSQLStorage(db *sql.DB, logger *logger.Logger) storage.Repository {
	s := &SQLStorage{
		db:     db,
		logger: logger,
	}

	return s
}

// SaveURL stores a new URL record in the SQL database.
func (s *SQLStorage) SaveURL(ctx context.Context, url entity.URLRecord) error {
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

// SaveURLBatch stores a batch of URL records in the SQL database in a single transaction.
func (s *SQLStorage) SaveURLBatch(ctx context.Context, urls []*entity.URLRecord) error {
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

// GetURL retrieves the original URL from the SQL database given its shortened version,
// unless it has been deleted.
func (s *SQLStorage) GetURL(ctx context.Context, shortURL string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	query := `
		SELECT original_url, is_deleted
		FROM short_urls
		WHERE short_url = $1`

	row := s.db.QueryRowContext(timeoutCtx, query, shortURL)

	var originalURL string
	var isDeleted bool
	if err := row.Scan(&originalURL, &isDeleted); err != nil {
		return "", err
	}

	if err := row.Err(); err != nil {
		return "", err
	}

	if isDeleted {
		return "", ErrURLDeleted
	}

	return originalURL, nil
}

// GetURLsByUserID retrieves all the URL records of a specific user from the SQL database.
func (s *SQLStorage) GetURLsByUserID(ctx context.Context, userID string) ([]entity.URLRecord, error) {
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

	var records []entity.URLRecord
	for rows.Next() {
		var r entity.URLRecord
		if err := rows.Scan(&r.UUID, &r.UserID, &r.ShortURL, &r.OriginalURL); err != nil {
			return nil, err
		}

		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no short urls for id %s", userID)
	}

	return records, nil
}

// CheckExistence checks if a shortened URL associated with a user exists in the SQL database.
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

// DeleteURLBatch marks a set of URLs associated with a user as deleted in SQL database
// by setting 'is_deleted' field to true for matching URLs.
func (s *SQLStorage) DeleteURLBatch(urls []string, user string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	urlsString := "{" + strings.Join(urls, ",") + "}"

	query := `
		UPDATE short_urls 
		SET is_deleted = true 
		WHERE short_url = ANY($1::text[]) AND user_id = $2::uuid`

	_, err := s.db.ExecContext(ctx, query, urlsString, user)
	if err != nil {
		return err
	}

	return nil
}

// Ping pings the database to check if it's alive.
func (s *SQLStorage) Ping() error {
	return s.db.Ping()
}

// OpenDB opens a SQL database connection given a DSN string.
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

// CreateTable creates a 'short_urls' table in the SQL database if it doesn't exist.
func CreateTable(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	query := `
		CREATE TABLE IF NOT EXISTS short_urls (
			id UUID UNIQUE,
			user_id UUID,
			short_url VARCHAR,
  			original_url VARCHAR,
  			is_deleted BOOLEAN DEFAULT false,
  			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
