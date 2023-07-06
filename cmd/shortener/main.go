package main

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/database"
	"github.com/PrahaTurbo/url-shortener/internal/storage/memory"
	"go.uber.org/zap"
	"log"
	"net/http"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	c := cfg.Load()
	if err := logger.Initialize(c.LogLevel); err != nil {
		log.Fatal(err)
	}

	var storage storage.Repository

	if c.DatabaseDSN == "" {
		storage = memory.NewStorage(c.StorageFilePath)
	} else {
		db, err := database.OpenDB(c.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		storage = database.NewStorage(db)
	}

	service := service.NewService(storage)
	app := app.NewApp(c, service)

	logger.Log.Info("Server is running", zap.String("address", app.Addr()))
	if err := http.ListenAndServe(app.Addr(), app.Router()); err != nil {
		log.Fatal(err)
	}
}
