package app

import (
	"github.com/go-chi/chi/v5"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
)

type App interface {
	Addr() string
	Router() chi.Router
}

type application struct {
	addr      string
	srv       service.Service
	logger    *logger.Logger
	jwtSecret string
}

func NewApp(addr, jwtSecret string, srv service.Service, logger *logger.Logger) App {
	return &application{
		addr:      addr,
		srv:       srv,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

func (a *application) Addr() string {
	return a.addr
}
