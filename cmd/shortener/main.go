package main

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
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

	log.Println("Running server on: ", app.Addr)
	if err := http.ListenAndServe(app.Addr, app.Router()); err != nil {
		log.Fatal(err)
	}
}
