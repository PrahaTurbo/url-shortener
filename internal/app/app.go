package app

import (
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type App interface {
	Addr() string
	Router() chi.Router
}

type application struct {
	addr   string
	srv    service.Service
	logger *logger.Logger
}

func NewApp(addr string, srv service.Service, logger *logger.Logger) App {
	return &application{
		addr:   addr,
		srv:    srv,
		logger: logger,
	}
}

func (a *application) Addr() string {
	return a.addr
}
