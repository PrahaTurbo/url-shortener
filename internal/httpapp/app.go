package httpapp

import (
	"github.com/PrahaTurbo/url-shortener/internal/auth"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
)

// Application is an implementation of the App interface.
type Application struct {
	srv    service.Service
	logger *logger.Logger
	auth   *auth.Auth
}

// NewHTTPApp initializes a new Application struct with the provided service, logger, server address and JWT Secret,
// and returns it as an App interface.
func NewHTTPApp(srv service.Service, logger *logger.Logger, auth *auth.Auth) *Application {
	return &Application{
		srv:    srv,
		logger: logger,
		auth:   auth,
	}
}
