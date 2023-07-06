package database

import (
	"database/sql"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"sync"
)

type SqlStorage struct {
	db *sql.DB
	mu sync.Mutex
}

func NewStorage(db *sql.DB) storage.Repository {
	return &SqlStorage{db: db}
}

func (s *SqlStorage) Put(id string, url []byte) {

}

func (s *SqlStorage) Get(id string) ([]byte, error) {
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

func (s *SqlStorage) Ping() error {
	return s.db.Ping()
}
