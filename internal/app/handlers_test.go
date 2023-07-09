package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/mock"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var baseURL = "http://localhost:8080"

func setupTestApp(mockStorage *mock.MockRepository) application {
	cfg := config.Config{
		Addr:    "localhost:8080",
		BaseURL: baseURL,
	}
	srv := service.NewService(cfg, mockStorage)

	return application{
		addr: cfg.Addr,
		srv:  srv,
	}
}

func Test_application_makeURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	urlRecord := storage.URLRecord{
		UUID:        "86d0f933-287c-4e1a-9978-4d9706e3e94f",
		ShortURL:    "fpCk-c",
		OriginalURL: "https://ya.ru",
	}

	s.EXPECT().
		GetURL(urlRecord.ShortURL).
		Return(&urlRecord, nil)

	s.EXPECT().
		GetURL("FgAJzm").
		Return(nil, errors.New("no url"))

	s.EXPECT().
		PutURL(gomock.Any()).
		Return(nil)

	app := setupTestApp(s)

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
			name:        "save url that already in storage",
			request:     "/",
			requestBody: urlRecord.OriginalURL,
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusConflict,
				response:    baseURL + "/" + urlRecord.ShortURL,
			},
		},
		{
			name:        "save new url",
			request:     "/",
			requestBody: "https://yandex.ru",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				response:    baseURL + "/FgAJzm",
			},
		},
		{
			name:        "unsupported request path",
			request:     "/make-url",
			requestBody: urlRecord.OriginalURL,
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
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	urlRecord := storage.URLRecord{
		UUID:        "86d0f933-287c-4e1a-9978-4d9706e3e94f",
		ShortURL:    "fpCk-c",
		OriginalURL: "https://ya.ru",
	}

	s.EXPECT().
		GetURL(urlRecord.ShortURL).
		Return(&urlRecord, nil)

	s.EXPECT().
		GetURL("Azcxc").
		Return(nil, errors.New("no url"))

	app := setupTestApp(s)

	type want struct {
		location   string
		statusCode int
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "success",
			request: "/" + urlRecord.ShortURL,
			want: want{
				location:   urlRecord.OriginalURL,
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name:    "wrong url id",
			request: "/Azcxc",
			want: want{
				location:   "",
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
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	urlRecord := storage.URLRecord{
		UUID:        "86d0f933-287c-4e1a-9978-4d9706e3e94f",
		ShortURL:    "fpCk-c",
		OriginalURL: "https://ya.ru",
	}

	s.EXPECT().
		GetURL(urlRecord.ShortURL).
		Return(&urlRecord, nil)

	s.EXPECT().
		GetURL("FgAJzm").
		Return(nil, errors.New("no url"))

	s.EXPECT().
		PutURL(gomock.Any()).
		Return(nil)

	app := setupTestApp(s)

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
			name:        "save url that already in storage",
			request:     "/api/shorten",
			requestBody: fmt.Sprintf(`{"url": "%s"}`, urlRecord.OriginalURL),
			want: want{
				statusCode: http.StatusConflict,
				response:   fmt.Sprintf(`{"result": "%s/%s"}`, baseURL, urlRecord.ShortURL),
			},
		},
		{
			name:        "save new url",
			request:     "/api/shorten",
			requestBody: `{"url": "https://yandex.ru"}`,
			want: want{
				statusCode: http.StatusCreated,
				response:   fmt.Sprintf(`{"result": "%s/FgAJzm"}`, baseURL),
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

func Test_application_pingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	tests := []struct {
		name       string
		statusCode int
		fail       bool
	}{
		{
			name:       "ping success",
			statusCode: 200,
			fail:       false,
		},
		{
			name:       "ping failed",
			statusCode: 500,
			fail:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			if tt.fail {
				s.EXPECT().
					Ping().
					Return(errors.New("no sql database"))
			} else {
				s.EXPECT().
					Ping().
					Return(nil)
			}

			app := setupTestApp(s)

			app.pingHandler(w, r)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func Test_application_batchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	s.EXPECT().
		GetURL(gomock.Any()).
		Return(nil, errors.New("no url")).AnyTimes()

	s.EXPECT().PutBatchURLs(gomock.Any()).
		Return(nil).AnyTimes()

	app := setupTestApp(s)

	successBody := fmt.Sprintf(`[{"correlation_id": "1", "short_url": "%s/fpCk-c"}]`, baseURL)

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
			request:     "/api/shorten/batch",
			requestBody: `[{"correlation_id": "1", "original_url": "https://ya.ru"}]`,
			want: want{
				statusCode: http.StatusCreated,
				response:   successBody,
			},
		},
		{
			name:        "post unmarshal error",
			request:     "/api/shorten/batch",
			requestBody: `[{"correlation_id": "1", "url": "https://ya.ru"}]`,
			want: want{
				statusCode: http.StatusInternalServerError,
				response:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, tt.request, reader)
			w := httptest.NewRecorder()
			app.batchHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)

			if tt.want.response != "" {
				assert.JSONEq(t, tt.want.response, w.Body.String())
			}
		})
	}
}
