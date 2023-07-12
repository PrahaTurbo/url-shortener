package app

import (
	"encoding/json"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/service"
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

	var statusCode int
	shortURL, err := a.srv.SaveURL(r.Context(), string(body))
	switch err {
	case service.ErrAlready:
		statusCode = http.StatusConflict
	case nil:
		statusCode = http.StatusCreated
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(shortURL))
}

func (a *application) jsonHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Debug("cannot unmarshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if req.URL == "" {
		a.logger.Debug("request without url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var statusCode int
	shortURL, err := a.srv.SaveURL(r.Context(), req.URL)
	switch err {
	case service.ErrAlready:
		statusCode = http.StatusConflict
	case nil:
		statusCode = http.StatusCreated
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)

	resp := models.Response{Result: shortURL}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		a.logger.Debug("error encoding response", zap.Error(err))
		return
	}
	a.logger.Debug("sending HTTP 201 response")
}

func (a *application) getOriginHandler(w http.ResponseWriter, r *http.Request) {
	url, err := a.srv.GetURL(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *application) pingHandler(w http.ResponseWriter, _ *http.Request) {
	if err := a.srv.PingDB(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *application) batchHandler(w http.ResponseWriter, r *http.Request) {
	var req []models.BatchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Debug("cannot unmarshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := a.srv.SaveBatch(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		a.logger.Debug("error encoding response", zap.Error(err))
		return
	}

	a.logger.Debug("sending HTTP 201 response")
}
