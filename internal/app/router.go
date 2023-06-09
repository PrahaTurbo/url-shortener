package app

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *application) router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Logger)

	r.Post("/", a.makeURLHandler)
	r.Get("/{id}", a.getOriginHandler)

	return r
}
