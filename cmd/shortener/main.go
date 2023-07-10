package main

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/memory"
	"github.com/PrahaTurbo/url-shortener/internal/storage/pg"
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

	var store storage.Repository

	if c.DatabaseDSN == "" {
		store = memory.NewInMemStorage(c.StorageFilePath)
	} else {
		db, err := pg.OpenDB(c.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		store = pg.NewSQLStorage(db)
	}

	srv := service.NewService(c, store)
	application := app.NewApp(c, srv)

	logger.Log.Info("Server is running", zap.String("address", application.Addr()))
	if err := http.ListenAndServe(application.Addr(), application.Router()); err != nil {
		log.Fatal(err)
	}
}
