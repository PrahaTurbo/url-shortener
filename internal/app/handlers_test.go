package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/mocks"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/PrahaTurbo/url-shortener/internal/storage/pg"
)

var (
	baseURL = "http://localhost:8080"
	addr    = "localhost:8080"
)

func setupTestApp() application {
	log, _ := logger.Initialize("debug")

	return application{
		addr:      addr,
		logger:    log,
		jwtSecret: "secret_key",
	}
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
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "shouldn't save url that already in storage",
			request:     "/",
			requestBody: "https://ya.ru",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveURL(gomock.Any(), "https://ya.ru").
					Return(baseURL+"/fpCk-c", service.ErrAlready)
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusConflict,
				response:    baseURL + "/fpCk-c",
			},
		},
		{
			name:        "should save new url successfully",
			request:     "/",
			requestBody: "https://ya.ru",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveURL(gomock.Any(), "https://ya.ru").
					Return(baseURL+"/fpCk-c", nil)
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				response:    baseURL + "/fpCk-c",
			},
		},
		{
			name:        "should return error if unsupported request path",
			request:     "/make-url",
			requestBody: "https://ya.ru",
			prepare:     func(s *mocks.MockService) {},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:        "should return internal server error",
			request:     "/",
			requestBody: "https://ya.ru",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveURL(gomock.Any(), "https://ya.ru").
					Return("", errors.New("internal error"))
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.srv = service

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
		prepare func(s *mocks.MockService)
		want    want
	}{
		{
			name:    "should successfully get original url",
			request: "/fpCk-c",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetURL(gomock.Any(), "fpCk-c").
					Return("https://ya.ru", nil)
			},
			want: want{
				location:   "https://ya.ru",
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name:    "should fail if wrong url id",
			request: "/azcxc",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetURL(gomock.Any(), "azcxc").
					Return("", errors.New("no url"))
			},
			want: want{
				location:   "",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "should return 410 if urls was deleted",
			request: "/fpCk-c",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetURL(gomock.Any(), "fpCk-c").
					Return("", pg.ErrURLDeleted)
			},
			want: want{
				location:   "",
				statusCode: http.StatusGone,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.srv = service

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
	app := setupTestApp()

	type want struct {
		statusCode int
		response   string
	}

	tests := []struct {
		name        string
		request     string
		requestBody string
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "shouldn't save url that already in storage",
			request:     "/api/shorten",
			requestBody: `{"url": "https://ya.ru"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveURL(gomock.Any(), "https://ya.ru").
					Return(baseURL+"/fpCk-c", service.ErrAlready)
			},
			want: want{
				statusCode: http.StatusConflict,
				response:   fmt.Sprintf(`{"result": "%s/fpCk-c"}`, baseURL),
			},
		},
		{
			name:        "should save new url successfully",
			request:     "/api/shorten",
			requestBody: `{"url": "https://yandex.ru"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveURL(gomock.Any(), "https://yandex.ru").
					Return(baseURL+"/FgAJzm", nil)
			},
			want: want{
				statusCode: http.StatusCreated,
				response:   fmt.Sprintf(`{"result": "%s/FgAJzm"}`, baseURL),
			},
		},
		{
			name:        "should fail with unmarshal error",
			request:     "/api/shorten",
			requestBody: `{"urm": "https://yandex.ru"}`,
			prepare:     func(s *mocks.MockService) {},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "should return internal server error",
			request:     "/api/shorten",
			requestBody: `{"url": "https://ya.ru"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveURL(gomock.Any(), "https://ya.ru").
					Return("", errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.srv = service

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
	app := setupTestApp()

	tests := []struct {
		name       string
		statusCode int
		prepare    func(s *mocks.MockService)
	}{
		{
			name:       "should ping successfully",
			statusCode: 200,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					PingDB().
					Return(nil)
			},
		},
		{
			name:       "should ping unsuccessfully",
			statusCode: 500,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					PingDB().
					Return(errors.New("no db"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.srv = service

			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			app.pingHandler(w, r)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func Test_application_batchHandler(t *testing.T) {
	app := setupTestApp()

	successBody := fmt.Sprintf(`[{"correlation_id": "1", "short_url": "%s/fpCk-c"}]`, baseURL)

	type want struct {
		statusCode int
		response   string
	}

	tests := []struct {
		name        string
		request     string
		requestBody string
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "should save batch urls successfully",
			request:     "/api/shorten/batch",
			requestBody: `[{"correlation_id": "1", "original_url": "https://ya.ru"}]`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveBatch(gomock.Any(), []models.BatchRequest{
						{
							CorrelationID: "1",
							OriginalURL:   "https://ya.ru",
						},
					}).
					Return([]models.BatchResponse{
						{
							CorrelationID: "1",
							ShortURL:      baseURL + "/fpCk-c",
						},
					}, nil)
			},
			want: want{
				statusCode: http.StatusCreated,
				response:   successBody,
			},
		},
		{
			name:        "should fail to save",
			request:     "/api/shorten/batch",
			requestBody: `[{"correlation_id": "1", "original_url": "https://ya.ru"}]`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					SaveBatch(gomock.Any(), []models.BatchRequest{
						{
							CorrelationID: "1",
							OriginalURL:   "https://ya.ru",
						},
					}).
					Return(nil, errors.New("can't save"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.srv = service

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

func Test_application_getUserURLsHandler(t *testing.T) {
	app := setupTestApp()

	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name    string
		request string
		prepare func(s *mocks.MockService)
		want    want
	}{
		{
			name:    "should return user urls successfully",
			request: "/api/user/urls",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetURLsByUserID(gomock.Any()).
					Return([]models.UserURLsResponse{
						{
							ShortURL:    baseURL + "/123",
							OriginalURL: "ya.ru",
						},
					}, nil)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name:    "should return 204 if user doesn't have urls",
			request: "/api/user/urls",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetURLsByUserID(gomock.Any()).
					Return(nil, errors.New("no urls"))
			},
			want: want{
				statusCode: http.StatusNoContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.srv = service

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)

			w := httptest.NewRecorder()
			app.getUserURLsHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-type"))
		})
	}
}

func Test_application_deleteURLsHandler(t *testing.T) {
	app := setupTestApp()

	type want struct {
		statusCode int
	}

	tests := []struct {
		name        string
		request     string
		requestBody string
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "should delete user urls successfully",
			request:     "/api/user/urls",
			requestBody: `["6qxTVvsy", "RTfd56hn", "Jlfd67ds"]`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					DeleteURLs(gomock.Any(), []string{"6qxTVvsy", "RTfd56hn", "Jlfd67ds"}).
					Return(nil)
			},
			want: want{
				statusCode: http.StatusAccepted,
			},
		},
		{
			name:    "should return error if cannot unmarshal",
			request: "/api/user/urls",
			prepare: func(s *mocks.MockService) {},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.srv = service

			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodDelete, tt.request, reader)

			w := httptest.NewRecorder()
			app.deleteURLsHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}
