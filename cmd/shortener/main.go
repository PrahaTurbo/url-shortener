package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/PrahaTurbo/url-shortener/internal/app"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	app := app.NewApp()

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
