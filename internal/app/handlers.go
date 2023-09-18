package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/PrahaTurbo/url-shortener/internal/storage/pg"
)

// makeURLHandler is an HTTP handler that saves URL from the request body and creates a short URL version.
// It responds with status codes to indicate success (201), duplicate URL (409),
// invalid path (400), or server errors (500).
//
// On successful URL creation, it returns the short URL in the response.
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

// jsonHandler is an HTTP handler that saves URL from the JSON in request body and creates a short URL version.
// It responds with status codes to indicate success (201), a URL already saved (409),
// a bad request i.e., a request without a URL (400), or server errors (500).
//
// On successful URL creation, it returns the short URL in the JSON response.
func (a *application) jsonHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Debug("cannot unmarshal request", zap.Error(err))
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	a.logger.Debug("sending HTTP 201 response")
}

// getOriginHandler is an HTTP handler function that retrieves the original URL
// for a given id from the request parameters.
// It responds with status codes to indicate success (307) alongside the original URL located in the header,
// a URL was deleted (410) or a bad request (400) for any other error.
func (a *application) getOriginHandler(w http.ResponseWriter, r *http.Request) {
	url, err := a.srv.GetURL(r.Context(), chi.URLParam(r, "id"))

	switch err {
	case pg.ErrURLDeleted:
		w.WriteHeader(http.StatusGone)
		return
	case nil:
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// pingHandler is an HTTP handler function that checks the connection to the database.
// It responds with status code 500 to indicate if the database is unreachable,
// or 200 if the connection is healthy.
func (a *application) pingHandler(w http.ResponseWriter, _ *http.Request) {
	if err := a.srv.PingDB(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// batchHandler is an HTTP handler that saves multiple URLs from the JSON in request body
// and creates a short URL version for each.
// It responds with status codes to indicate success (201), or server errors (500).
//
// On successful URLs creation, it returns the short URLs in the JSON response.
func (a *application) batchHandler(w http.ResponseWriter, r *http.Request) {
	var req []models.BatchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Debug("cannot unmarshal request", zap.Error(err))
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.logger.Debug("sending HTTP 201 response")
}

// getUserURLsHandler is an HTTP handler function that retrieves all URLs associated
// with a user ID from the request context.
// It responds with status codes to indicate success (200), if no URLs are found (204),
// or server errors (500).
//
// On success, it returns the short URLs in the JSON response.
func (a *application) getUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := a.srv.GetURLsByUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("content-type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		a.logger.Debug("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.logger.Debug("sending HTTP 200 response")
}

// deleteURLsHandler is an HTTP handler function that deletes all URLs
// associated with the user.
// It responds with status codes to indicate when request accepted (202),or server errors (500).
func (a *application) deleteURLsHandler(w http.ResponseWriter, r *http.Request) {
	var shortURLs []string

	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		a.logger.Debug("cannot decode request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	if err := a.srv.DeleteURLs(r.Context(), shortURLs); err != nil {
		a.logger.Debug("cannot delete short urls for user", zap.Error(err))
	}
}
