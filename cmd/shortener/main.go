package main

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/app"
)

func main() {
	c := cfg.Load()
	if err := logger.Initialize(c.LogLevel); err != nil {
		log.Fatal(err)
	}
	app := app.NewApp(c)

	logger.Log.Info("Server is running", zap.String("address", app.Addr))
	if err := http.ListenAndServe(app.Addr, app.Router()); err != nil {
		log.Fatal(err)
	}
}
