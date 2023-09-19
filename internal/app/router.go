package app

import (
	"github.com/go-chi/chi/v5"
	libmiddleware "github.com/go-chi/chi/v5/middleware"

	appmiddleware "github.com/PrahaTurbo/url-shortener/internal/middleware"
)

func (a *Application) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(a.logger.RequestLogger)
	r.Use(appmiddleware.Auth(a.jwtSecret))
	r.Use(libmiddleware.Compress(5, "application/json", "text/html"))
	r.Use(appmiddleware.Decompress)
	r.Mount("/debug", libmiddleware.Profiler())

	r.Post("/", a.MakeURLHandler)
	r.Get("/{id}", a.GetOriginHandler)
	r.Post("/api/shorten", a.JSONHandler)
	r.Post("/api/shorten/batch", a.BatchHandler)
	r.Get("/api/user/urls", a.GetUserURLsHandler)
	r.Delete("/api/user/urls", a.DeleteURLsHandler)
	r.Get("/ping", a.PingHandler)

	return r
}
