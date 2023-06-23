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
		Addr:    "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	app := application{
		Addr:    cfg.Addr,
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
		response    string
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
				response:    app.baseURL + "/FgAJzm",
			},
		},
		{
			name:        "unsupported request path",
			request:     "/make-url",
			requestBody: "https://yandex.ru",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
				response:    "",
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
			assert.Equal(t, tt.want.response, w.Body.String())
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

func Test_application_jsonHandler(t *testing.T) {
	app := setupTestApp()

	successBody := fmt.Sprintf(`{"result": "http://%s/FgAJzm"}`, app.Addr)

	type want struct {
		statusCode int
		response   string
	}

	tests := []struct {
		name        string
		request     string
		requestBody string
		want        want
	}{
		{
			name:        "post success",
			request:     "/api/shorten",
			requestBody: `{"url": "https://yandex.ru"}`,
			want: want{
				statusCode: http.StatusOK,
				response:   successBody,
			},
		},
		{
			name:        "post unmarshal error",
			request:     "/api/shorten",
			requestBody: `{"urm": "https://yandex.ru"}`,
			want: want{
				statusCode: http.StatusBadRequest,
				response:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, tt.request, reader)
			w := httptest.NewRecorder()
			app.jsonHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)

			if tt.want.response != "" {
				assert.JSONEq(t, tt.want.response, w.Body.String())
			}
		})
	}
}
