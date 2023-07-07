package database

import (
	"context"
	"database/sql"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"go.uber.org/zap"
	"sync"
	"time"
)

type SQLStorage struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLStorage(db *sql.DB) storage.Repository {
	createTable(db)

	return &SQLStorage{db: db}
}

func (s *SQLStorage) Put(id string, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	stmt := `
INSERT INTO short_urls (id, original_url)
VALUES ($1, $2)`

	res, err := s.db.ExecContext(ctx, stmt, id, url)
	if err != nil {
		logger.Log.Error("cannot save url to db", zap.Error(err))
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("error while getting affected rows", zap.Error(err))
		return
	}

	if affectedRows != 1 {
		logger.Log.Error("no rows was affected by instert in database")
		return
	}
}

func (s *SQLStorage) Get(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	stmt := `
SELECT original_url 
FROM short_urls
WHERE id = $1`

	row := s.db.QueryRowContext(ctx, stmt, id)

	var url string
	if err := row.Scan(&url); err != nil {
		return "", err
	}

	if err := row.Err(); err != nil {
		return "", err
	}

	return url, nil
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
    id VARCHAR PRIMARY KEY, 
    original_url VARCHAR)`

	_, err := db.Exec(stmt)
	if err != nil {
		logger.Log.Fatal("cannot create table in database", zap.Error(err))
	}
}
