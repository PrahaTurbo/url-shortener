package database

import (
	"database/sql"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"sync"
)

type SQLStorage struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLStorage(db *sql.DB) storage.Repository {
	return &SQLStorage{db: db}
}

func (s *SQLStorage) Put(id string, url []byte) {

}

func (s *SQLStorage) Get(id string) ([]byte, error) {
	return nil, nil
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
