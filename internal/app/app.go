package app

import (
	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/service"
)

type application struct {
	Addr    string
	baseURL string
	srv     service.Service
}

func NewApp(c cfg.Config) application {
	return application{
		Addr:    c.Addr,
		baseURL: c.BaseURL,
		srv:     service.NewService(),
	}
}
