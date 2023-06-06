package app

import (
	"log"
	"net/http"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	s "github.com/PrahaTurbo/url-shortener/internal/storage"
)

type application struct {
	addr    string
	baseURL string
	srv     service.Service
}

func NewApp(c cfg.Config) application {
	return application{
		addr:    c.Host + ":" + c.Port,
		baseURL: c.BaseURL,
		srv: service.Service{
			URLs: &s.Storage{DB: make(map[string][]byte)},
		},
	}
}

func (a *application) Start() error {
	log.Println("Running server on: ", a.addr)
	return http.ListenAndServe(a.addr, a.router())
}
