package pg

import (
	"context"
	"database/sql"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"go.uber.org/zap"
	"time"
)

type SQLStorage struct {
	db *sql.DB
}

func NewSQLStorage(db *sql.DB) storage.Repository {
	createTable(db)

	return &SQLStorage{db: db}
}

func (s *SQLStorage) PutURL(url storage.URLRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	stmt := `
		INSERT INTO short_urls (id, short_url, original_url)
		VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, stmt, url.UUID, url.ShortURL, url.OriginalURL)
	if err != nil {
		logger.Log.Error("cannot instert into", zap.Error(err))
		return err
	}

	return nil
}

func (s *SQLStorage) PutBatchURLs(urls []storage.URLRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO short_urls (id, short_url, original_url)
		VALUES ($1, $2, $3)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err := stmt.ExecContext(ctx, url.UUID, url.ShortURL, url.OriginalURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLStorage) GetURL(shortUrl string) (*storage.URLRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	stmt := `
		SELECT id, short_url, original_url 
		FROM short_urls
		WHERE short_url = $1`

	row := s.db.QueryRowContext(ctx, stmt, shortUrl)

	var url storage.URLRecord
	if err := row.Scan(&url.UUID, &url.ShortURL, &url.OriginalURL); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return &url, nil
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

func (s *SQLStorage) Ping() error {
	return s.db.Ping()
}

func createTable(db *sql.DB) {
	stmt := `
		CREATE TABLE IF NOT EXISTS short_urls (
			id UUID,
			short_url VARCHAR PRIMARY KEY,
  			original_url VARCHAR,
  			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`

	_, err := db.Exec(stmt)
	if err != nil {
		logger.Log.Fatal("cannot create table in database", zap.Error(err))
	}
}
