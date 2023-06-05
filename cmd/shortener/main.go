package main

import (
	"log"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/app"
)

func main() {
	c := cfg.New()
	app := app.NewApp(c)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
