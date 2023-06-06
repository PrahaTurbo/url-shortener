package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type storageMock struct {
	db map[string][]byte
}

func (s *storageMock) Put(id string, url []byte) {
	s.db[id] = url
}

func (s *storageMock) Get(id string) ([]byte, error) {
	url, ok := s.db[id]
	if !ok {
		return nil, fmt.Errorf("no url for id: %s", id)
	}

	return url, nil
}

func setupTestApp() application {
	cfg := config.Config{
		Host:    "localhost",
		Port:    "8080",
		BaseURL: "localhost:8080",
	}
	app := application{
		addr:    cfg.Host + ":" + cfg.Port,
		baseURL: cfg.BaseURL,
		srv: service.Service{
			URLs: &storageMock{db: make(map[string][]byte)},
		},
	}

	return app
}

func Test_application_makeURL(t *testing.T) {
	app := setupTestApp()

	type want struct {
		contentType string
		statusCode  int
		reponse     string
	}

	tests := []struct {
		name        string
		request     string
		requestBody string
		want        want
	}{
		{
			name:        "simple url",
			request:     "/",
			requestBody: "https://yandex.ru",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				reponse:     fmt.Sprintf("%s/%s", app.baseURL, app.srv.SaveURL([]byte("https://yandex.ru"))),
			},
		},
		{
			name:        "unsupported request path",
			request:     "/make-url",
			requestBody: "https://yandex.ru",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
				reponse:     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, tt.request, reader)
			w := httptest.NewRecorder()
			app.makeURLHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)

			if w.Code != http.StatusCreated {
				return
			}

			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-type"))
			assert.Equal(t, tt.want.reponse, w.Body.String())
		})
	}
}

func Test_application_getOrigin(t *testing.T) {
	app := setupTestApp()

	type want struct {
		location   string
		statusCode int
	}

	tests := []struct {
		name    string
		request string
		urlID   string
		want    want
	}{
		{
			name:    "success",
			request: "/yAvfdS",
			urlID:   "yAvfdS",
			want: want{
				location:   "https://yandex.ru",
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name:    "wrong url id",
			request: "/yAvFbv",
			urlID:   "Azcxc",
			want: want{
				location:   "https://yandex.ru",
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
			chiCtx.URLParams.Add("id", tt.request[1:])

			app.srv.URLs.Put(tt.urlID, []byte(tt.want.location))
			app.getOriginHandler(w, r)

			assert.Equal(t, tt.want.statusCode, w.Code)

			if w.Code != http.StatusTemporaryRedirect {
				return
			}

			assert.Equal(t, tt.want.location, w.Header().Get("Location"))
		})
	}
}
