package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/mocks"
	"github.com/PrahaTurbo/url-shortener/internal/models"
)

var baseURL = "localhost:8080"

func setupService() service {
	log, _ := logger.Initialize("debug")

	return service{
		baseURL: baseURL,
		logger:  log,
	}
}

func TestService_SaveURL(t *testing.T) {
	service := setupService()

	type want struct {
		url string
		err error
	}

	tests := []struct {
		name    string
		url     string
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should save url successfully",
			url:  "https://yandex.ru",
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						CheckExistence(gomock.Any(), "FgAJzm", "1").
						Return(errors.New("no url")),
					s.EXPECT().
						SaveURL(gomock.Any(), gomock.Any()).
						Return(nil),
				)
			},
			want: want{
				url: baseURL + "/" + "FgAJzm",
			},
		},
		{
			name: "shouldn't save url that already in storage",
			url:  "https://yandex.ru",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					CheckExistence(gomock.Any(), "FgAJzm", "1").
					Return(nil)
			},
			want: want{
				url: baseURL + "/" + "FgAJzm",
				err: ErrAlready,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)
			ctx := context.WithValue(context.Background(), config.UserIDKey, "1")

			tt.prepare(storage)
			service.Storage = storage

			shortURL, err := service.SaveURL(ctx, tt.url)

			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.url, shortURL)
		})
	}
}

func TestService_GetURL(t *testing.T) {
	service := setupService()

	type want struct {
		url string
		err bool
	}

	tests := []struct {
		name     string
		shortURL string
		prepare  func(s *mocks.MockRepository)
		want     want
	}{
		{
			name:     "should get origin url successfully",
			shortURL: "fpCk-c",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetURL(gomock.Any(), "fpCk-c").
					Return("https://ya.ru", nil)
			},
			want: want{
				url: "https://ya.ru",
				err: false,
			},
		},
		{
			name:     "should fail when getting url with false id",
			shortURL: "abc",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetURL(gomock.Any(), "abc").
					Return("", errors.New("no url"))
			},
			want: want{
				url: "",
				err: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)
			ctx := context.WithValue(context.Background(), config.UserIDKey, "1")

			tt.prepare(storage)
			service.Storage = storage

			url, err := service.GetURL(ctx, tt.shortURL)
			if !tt.want.err {
				require.NoError(t, err)

				assert.Equal(t, url, tt.want.url)
				return
			}

			assert.Error(t, err)
		})
	}
}

func TestService_SaveBatch(t *testing.T) {
	service := setupService()

	batch := []models.BatchRequest{
		{
			CorrelationID: "1",
			OriginalURL:   "https://ya.ru",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://yandex.ru",
		},
	}

	type want struct {
		resp []models.BatchResponse
		err  bool
	}

	tests := []struct {
		name     string
		batchReq []models.BatchRequest
		prepare  func(s *mocks.MockRepository)
		want     want
	}{
		{
			name:     "should save batch successfully",
			batchReq: batch,
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					CheckExistence(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("no url")).AnyTimes()
				s.EXPECT().
					SaveURLBatch(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			want: want{
				resp: []models.BatchResponse{
					{
						CorrelationID: "1",
						ShortURL:      baseURL + "/fpCk-c",
					},
					{
						CorrelationID: "2",
						ShortURL:      baseURL + "/FgAJzm",
					},
				},
			},
		},
		{
			name:     "should fail while saving batch",
			batchReq: batch,
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					CheckExistence(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("no url")).AnyTimes()
				s.EXPECT().
					SaveURLBatch(gomock.Any(), gomock.Any()).
					Return(errors.New("cannot save batch urls"))
			},
			want: want{
				err: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)
			ctx := context.WithValue(context.Background(), config.UserIDKey, "1")

			tt.prepare(storage)
			service.Storage = storage

			resp, err := service.SaveBatch(ctx, tt.batchReq)

			if tt.want.err {
				assert.NotEmpty(t, err)
				return
			}

			if assert.NoError(t, err) {
				assert.Equal(t, tt.want.resp, resp)
			}
		})
	}
}

func BenchmarkService_SaveBatch(b *testing.B) {
	service := setupService()
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockRepository(ctrl)
	ctx := context.WithValue(context.Background(), config.UserIDKey, "1")

	storage.EXPECT().
		CheckExistence(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	storage.EXPECT().
		SaveURLBatch(gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	service.Storage = storage

	var batchReqs []models.BatchRequest
	for i := 0; i < len(batchReqs); i++ {
		batchReqs = append(batchReqs, models.BatchRequest{
			CorrelationID: "12344567",
			OriginalURL:   "yandex.ru",
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.SaveBatch(ctx, batchReqs)
	}
}

func BenchmarkService_DeleteURLs(b *testing.B) {
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockRepository(ctrl)
	log, _ := logger.Initialize("debug")
	ctx := context.WithValue(context.Background(), config.UserIDKey, "1")

	storage.EXPECT().
		DeleteURLBatch(gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	service := NewService(baseURL, storage, log)

	urls := []string{strings.Repeat("yandex.ru", 1000)}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.DeleteURLs(ctx, urls)
	}
}
