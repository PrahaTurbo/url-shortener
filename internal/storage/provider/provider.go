package provider

import (
	"log"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/memory"
	"github.com/PrahaTurbo/url-shortener/internal/storage/pg"
)

// NewStorage creates a new storage repository based on provided parameters.
// If a data source name (dsn) is not provided, it returns an in-memory storage that uses a file for persistence.
// If a dsn is provided, it opens a connection to a PostgreSQL database and returns a SQL-based storage repository.
// It also executes a method to create the necessary table in the PostgreSQL database if it does not already exist.
func NewStorage(dsn string, filePath string, logger *logger.Logger) (storage.Repository, error) {
	if dsn == "" {
		return memory.NewInMemStorage(filePath, logger), nil
	}

	db, err := pg.OpenDB(dsn)
	if err != nil {
		log.Fatal(err)
	}

	store := pg.NewSQLStorage(db, logger)
	if err := pg.CreateTable(db); err != nil {
		return nil, err
	}

	return store, nil
}
