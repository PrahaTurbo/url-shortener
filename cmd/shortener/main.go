package main

import (
	"log"
	"math/rand"
	"time"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/app"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	c := cfg.New()

	app := app.NewApp(c)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
