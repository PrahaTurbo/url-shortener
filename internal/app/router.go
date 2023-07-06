package app

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func (a *application) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	r.Use(middleware.Gzip)

	r.Post("/", a.makeURLHandler)
	r.Get("/{id}", a.getOriginHandler)
	r.Post("/api/shorten", a.jsonHandler)
	r.Get("/ping", a.pingHandler)

	return r
}
