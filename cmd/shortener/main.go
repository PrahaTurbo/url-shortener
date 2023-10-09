package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/app"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/PrahaTurbo/url-shortener/internal/storage/provider"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	c := cfg.Load()
	lgr, err := logger.Initialize(c.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	store, err := provider.NewStorage(c.DatabaseDSN, c.StorageFilePath, lgr)
	if err != nil {
		log.Fatal(err)
	}

	srv := service.NewService(c.BaseURL, store, lgr)
	application := app.NewApp(c.Addr, c.JWTSecret, srv, lgr)

	lgr.Info("Server is running", zap.String("address", application.Addr()))
	if err := http.ListenAndServe(application.Addr(), application.Router()); err != nil {
		log.Fatal(err)
	}
}
