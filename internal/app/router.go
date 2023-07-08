package app

import (
	"github.com/PrahaTurbo/url-shortener/internal/gzip"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *application) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	r.Use(middleware.Compress(5, "application/json", "text/html"))
	r.Use(gzip.Decompress)

	r.Post("/", a.makeURLHandler)
	r.Get("/{id}", a.getOriginHandler)
	r.Post("/api/shorten", a.jsonHandler)
	r.Get("/ping", a.pingHandler)

	return r
}
