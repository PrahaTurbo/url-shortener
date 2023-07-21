package app

import (
	appmiddleware "github.com/PrahaTurbo/url-shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	libmiddleware "github.com/go-chi/chi/v5/middleware"
)

func (a *application) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(a.logger.RequestLogger)
	r.Use(appmiddleware.Auth(a.jwtSecret))
	r.Use(libmiddleware.Compress(5, "application/json", "text/html"))
	r.Use(appmiddleware.Decompress)

	r.Post("/", a.makeURLHandler)
	r.Get("/{id}", a.getOriginHandler)
	r.Post("/api/shorten", a.jsonHandler)
	r.Post("/api/shorten/batch", a.batchHandler)
	r.Get("/api/user/urls", a.getUserURLsHandler)
	r.Get("/ping", a.pingHandler)

	return r
}
