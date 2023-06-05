package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_application_makeURL(t *testing.T) {
	cfg := config.Config{Host: "localhost", Port: "8080"}
	app := application{
		addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		srv: service.Service{
			URLs: &storageMock{db: make(map[string][]byte)},
		},
	}

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
				reponse:     fmt.Sprintf("http://%s/%s", app.addr, app.srv.SaveURL([]byte("https://yandex.ru"))),
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
			app.makeURL(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if result.StatusCode != http.StatusCreated {
				return
			}

			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-type"))

			resultURL, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.reponse, string(resultURL))
		})
	}
}

func Test_application_getOrigin(t *testing.T) {
	cfg := config.Config{Host: "localhost", Port: "8080"}
	app := application{
		addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		srv: service.Service{
			URLs: &storageMock{db: make(map[string][]byte)},
		},
	}

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
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			app.srv.URLs.Put(tt.urlID, []byte(tt.want.location))

			app.getOrigin(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if result.StatusCode != http.StatusTemporaryRedirect {
				return
			}

			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}

func Test_application_rootHandler(t *testing.T) {
	cfg := config.Config{Host: "localhost", Port: "8080"}
	app := application{
		addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		srv: service.Service{
			URLs: &storageMock{db: make(map[string][]byte)},
		},
	}
	app.srv.URLs.Put("id", []byte("site.com"))

	type want struct {
		statusCode int
	}

	tests := []struct {
		name          string
		requestMethod string
		request       string
		want          want
	}{
		{
			name:          "post routing",
			requestMethod: http.MethodPost,
			request:       "/",
			want: want{
				statusCode: http.StatusCreated,
			},
		},
		{
			name:          "get routing",
			requestMethod: http.MethodGet,
			request:       "/id",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name:          "wrong method",
			requestMethod: http.MethodPut,
			request:       "/",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.requestMethod, tt.request, nil)
			w := httptest.NewRecorder()

			app.rootHandler(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}
