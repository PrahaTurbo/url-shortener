package app

import (
	"fmt"
	"io"
	"net/http"
)

func (a *application) rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.makeURL(w, r)
	case http.MethodGet:
		a.getOrigin(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *application) makeURL(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	urlID := a.srv.SaveURL(body)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://%s/%s", a.addr, urlID)))
}

func (a *application) getOrigin(w http.ResponseWriter, r *http.Request) {
	url, err := a.srv.GetURL(r.URL.Path[1:])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", string(url))
	w.WriteHeader(http.StatusTemporaryRedirect)
}
