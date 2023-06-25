package app

import (
	"encoding/json"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *application) makeURLHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(a.baseURL + "/" + urlID))
}

func (a *application) jsonHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Debug("cannot unmarshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if req.URL == "" {
		logger.Log.Debug("request without url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlID := a.srv.SaveURL([]byte(req.URL))

	resp := models.Response{Result: a.baseURL + "/" + urlID}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
	logger.Log.Debug("sending HTTP 200 response")
}

func (a *application) getOriginHandler(w http.ResponseWriter, r *http.Request) {
	url, err := a.srv.GetURL(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", string(url))
	w.WriteHeader(http.StatusTemporaryRedirect)
}
