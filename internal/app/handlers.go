package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (a *application) Root(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.makeUrl(w, r)
	case http.MethodGet:
		a.getOrigin(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (a *application) makeUrl(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlID := genID(6)
	a.urls[urlID] = body

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://%s/%s", a.Addr, urlID)))
}

func (a *application) getOrigin(w http.ResponseWriter, r *http.Request) {
	url, ok := a.urls[strings.Trim(r.URL.Path, "/")]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", string(url))
	w.WriteHeader(http.StatusTemporaryRedirect)
}
