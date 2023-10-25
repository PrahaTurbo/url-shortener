package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	srvc := service.NewService(c.BaseURL, store, lgr)
	application := app.NewApp(c.Addr, c.JWTSecret, srvc, lgr)

	server := http.Server{
		Addr:    application.Addr(),
		Handler: application.Router(),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sigint

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		close(idleConnsClosed)
	}()

	lgr.Info("Server is running", zap.String("address", application.Addr()))

	if c.EnableHTTPS {
		if err := server.ListenAndServeTLS(
			"cmd/shortener/cert.pem",
			"cmd/shortener/key.pem",
		); !errors.Is(err, http.ErrServerClosed) {
			lgr.Fatal("url shortener server error", zap.Error(err))
		}
	} else {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			lgr.Fatal("url shortener server error", zap.Error(err))
		}
	}

	<-idleConnsClosed
	lgr.Info("Server Shutdown gracefully", zap.String("address", application.Addr()))
}
