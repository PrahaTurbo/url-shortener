package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/PrahaTurbo/url-shortener/internal/app"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	app := app.NewApp()

	http.HandleFunc("/", app.Root)

	if err := http.ListenAndServe(app.Addr, nil); err != nil {
		panic(err)
	}
}
