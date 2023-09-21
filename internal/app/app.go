package app

import (
	"github.com/go-chi/chi/v5"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
)

// App is an interface representing the web server of the application, which handles requests and responses.
// It provides public methods to access the server's address and router configuration.
type App interface {
	Router() chi.Router
	Addr() string
}

// Application is an implementation of the App interface.
type Application struct {
	addr      string
	srv       service.Service
	logger    *logger.Logger
	jwtSecret string
}

// NewApp initializes a new Application struct with the provided service, logger, server address and JWT Secret,
// and returns it as an App interface.
func NewApp(addr, jwtSecret string, srv service.Service, logger *logger.Logger) App {
	return &Application{
		addr:      addr,
		srv:       srv,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

// Addr is a receiver function on the Application struct that returns the server's address.
func (a *Application) Addr() string {
	return a.addr
}
