package app

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *application) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/", a.makeURLHandler)
	r.Get("/{id}", a.getOriginHandler)

	return r
}
