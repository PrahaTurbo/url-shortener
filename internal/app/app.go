package app

import (
	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/service"
)

type application struct {
	addr string
	srv  service.Service
}

func NewApp(c cfg.Config, srv service.Service) application {
	return application{
		addr: c.Addr,
		srv:  srv,
	}
}

func (a *application) Addr() string {
	return a.addr
}
