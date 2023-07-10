package app

import (
	"encoding/json"
	"errors"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
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
	shortURL, err := a.srv.SaveURL(string(body))
	if err != nil {
		if errors.Is(err, service.ErrAlready) {
			statusCode = http.StatusConflict
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		statusCode = http.StatusCreated
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(shortURL))
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

	var statusCode int
	shortURL, err := a.srv.SaveURL(req.URL)
	if err != nil {
		if errors.Is(err, service.ErrAlready) {
			statusCode = http.StatusConflict
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		statusCode = http.StatusCreated
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)

	resp := models.Response{Result: shortURL}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
	logger.Log.Debug("sending HTTP 201 response")
}

func (a *application) getOriginHandler(w http.ResponseWriter, r *http.Request) {
	url, err := a.srv.GetURL(chi.URLParam(r, "id"))
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
		logger.Log.Debug("cannot unmarshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := a.srv.SaveBatch(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}

	logger.Log.Debug("sending HTTP 201 response")
}
