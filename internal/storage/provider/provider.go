package provider

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/memory"
	"github.com/PrahaTurbo/url-shortener/internal/storage/pg"
	"log"
)

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
