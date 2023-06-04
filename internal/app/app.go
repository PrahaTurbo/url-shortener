package app

import (
	"fmt"
	"net/http"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	srv "github.com/PrahaTurbo/url-shortener/internal/service"
	s "github.com/PrahaTurbo/url-shortener/internal/storage"
)

type application struct {
	urls map[string][]byte
	addr string
	srv  srv.Service
}

func NewApp(c cfg.Config) application {
	return application{
		urls: make(map[string][]byte),
		addr: fmt.Sprintf("%s:%s", c.Host, c.Port),
		srv: srv.Service{
			DB: s.Storage{DB: make(map[string][]byte)},
		},
	}
}

func (a *application) Start() error {
	http.HandleFunc("/", a.rootHandler)

	return http.ListenAndServe(a.addr, nil)
}
