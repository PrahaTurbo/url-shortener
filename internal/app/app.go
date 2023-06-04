package app

import "net/http"

type application struct {
	urls map[string][]byte
	addr string
}

func NewApp() application {
	return application{
		urls: make(map[string][]byte),
		addr: "localhost:8080",
	}
}

func (a *application) Start() error {
	http.HandleFunc("/", a.rootHandler)

	return http.ListenAndServe(a.addr, nil)
}
