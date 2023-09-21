package app

import (
	"github.com/go-chi/chi/v5"
	libmiddleware "github.com/go-chi/chi/v5/middleware"

	appmiddleware "github.com/PrahaTurbo/url-shortener/internal/middleware"
)

// Router is a receiver method on the Application struct that initializes and returns a new chi Router.
// It sets up middleware functions for logging, authentication, compression and decompression.
// It also maps HTTP methods (GET, POST, DELETE) and routes to the appropriate handler functions.
func (a *Application) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(a.logger.RequestLogger)
	r.Use(appmiddleware.Auth(a.jwtSecret))
	r.Use(libmiddleware.Compress(5, "application/json", "text/html"))
	r.Use(appmiddleware.Decompress)

	r.Post("/", a.MakeURLHandler)
	r.Get("/{id}", a.GetOriginHandler)
	r.Post("/api/shorten", a.JSONHandler)
	r.Post("/api/shorten/batch", a.BatchHandler)
	r.Get("/api/user/urls", a.GetUserURLsHandler)
	r.Delete("/api/user/urls", a.DeleteURLsHandler)
	r.Get("/ping", a.PingHandler)

	return r
}
