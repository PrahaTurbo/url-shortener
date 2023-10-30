package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/PrahaTurbo/url-shortener/internal/auth"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/mocks"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/storage/entity"
)

type badContextKey string

var baseURL = "localhost:8080"
var errInternal = errors.New("internal error")

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
		{
			name:    "should return error if can't extract user id from context",
			url:     "https://yandex.ru",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
		{
			name: "should return error if failed to save url",
			url:  "https://yandex.ru",
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						CheckExistence(gomock.Any(), "FgAJzm", "1").
						Return(errors.New("no url")),
					s.EXPECT().
						SaveURL(gomock.Any(), gomock.Any()).
						Return(errInternal),
				)
			},
			want: want{
				err: errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)
			ctx := context.WithValue(context.Background(), auth.UserIDKey, "1")

			if errors.Is(tt.want.err, ErrExtractFromContext) {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

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
			ctx := context.WithValue(context.Background(), auth.UserIDKey, "1")

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
		err  error
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
			name:     "should skip one url if it is already in storage",
			batchReq: batch,
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					CheckExistence(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("no url"))
				s.EXPECT().
					CheckExistence(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				s.EXPECT().
					SaveURLBatch(gomock.Any(), gomock.Len(1)).
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
					Return(errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
		{
			name:     "should return error if can't extract user id from context",
			batchReq: batch,
			prepare:  func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
		{
			name: "should return error if there is no original_url in batch request",
			batchReq: []models.BatchRequest{
				{
					CorrelationID: "1",
				},
			},
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrNoOriginalURL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)
			ctx := context.WithValue(context.Background(), auth.UserIDKey, "1")

			if errors.Is(tt.want.err, ErrExtractFromContext) {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			tt.prepare(storage)
			service.Storage = storage

			resp, err := service.SaveBatch(ctx, tt.batchReq)

			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.resp, resp)
		})
	}
}

func TestService_GetURLsByUserID(t *testing.T) {
	service := setupService()

	type want struct {
		resp []models.UserURLsResponse
		err  error
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should save batch successfully",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetURLsByUserID(gomock.Any(), "1").
					Return([]entity.URLRecord{
						{
							ShortURL:    "123abc",
							OriginalURL: "ya.ru",
						},
					}, nil)
			},
			want: want{
				resp: []models.UserURLsResponse{
					{
						ShortURL:    baseURL + "/123abc",
						OriginalURL: "ya.ru",
					},
				},
			},
		},
		{
			name: "should fail while getting user urls",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetURLsByUserID(gomock.Any(), "1").
					Return(nil, errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
		{
			name:    "should return error if can't extract user id from context",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)
			ctx := context.WithValue(context.Background(), auth.UserIDKey, "1")

			if errors.Is(tt.want.err, ErrExtractFromContext) {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			tt.prepare(storage)
			service.Storage = storage

			resp, err := service.GetURLsByUserID(ctx)

			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.resp, resp)
		})
	}
}

func TestService_PingDB(t *testing.T) {
	service := setupService()

	tests := []struct {
		name    string
		prepare func(s *mocks.MockRepository)
		wantErr bool
	}{
		{
			name: "should ping  successfully",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().Ping().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "should ping failed",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().Ping().Return(errors.New("no db"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.Storage = storage

			err := service.PingDB()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestService_DeleteURLs(t *testing.T) {
	service := setupService()

	tests := []struct {
		name    string
		urls    []string
		wantErr error
	}{
		{
			name: "should delete urls successfully",
			urls: []string{strings.Repeat("yandex.ru", 10)},
		},
		{
			name:    "should return error if can't extract user id from context",
			wantErr: ErrExtractFromContext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), auth.UserIDKey, "1")

			if errors.Is(tt.wantErr, ErrExtractFromContext) {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			service.delChan = make(chan models.URLDeletionTask, 10)

			go service.startURLDeletionWorker(time.Second*1, 10)

			err := service.DeleteURLs(ctx, tt.urls)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_GetStats(t *testing.T) {
	service := setupService()

	type want struct {
		resp *models.StatsResponse
		err  error
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should get stats successfully",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetStats(gomock.Any()).
					Return(&entity.Stats{
						URLs:  13,
						Users: 10,
					}, nil)
			},
			want: want{
				resp: &models.StatsResponse{
					URLs:  13,
					Users: 10,
				},
			},
		},
		{
			name: "should fail while getting stats",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetStats(gomock.Any()).
					Return(nil, errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.Storage = storage

			resp, err := service.GetStats(context.Background())

			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.resp, resp)
		})
	}
}

func Test_service_handleDeletion(t *testing.T) {
	service := setupService()

	tasks := []models.URLDeletionTask{
		{
			UserID: "testuser1",
			URLs: []string{
				"http://test1.com",
				"http://test2.com",
			},
		},
		{
			UserID: "testuser2",
			URLs: []string{
				"http://test3.com",
				"http://test4.com",
			},
		},
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockRepository)
	}{
		{
			name: "should delete urls successfully",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					DeleteURLBatch(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(2)
			},
		},
		{
			name: "should delete one url successfully",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					DeleteURLBatch(gomock.Any(), gomock.Any()).
					Return(nil)
				s.EXPECT().
					DeleteURLBatch(gomock.Any(), gomock.Any()).
					Return(errInternal)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.Storage = storage
			service.semaphore = newSemaphore(5)

			service.handleDeletion(tasks)

			for {
				if len(service.semaphore.semaCh) == 0 {
					return
				}
			}
		})
	}
}

func BenchmarkService_SaveBatch(b *testing.B) {
	service := setupService()
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockRepository(ctrl)
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "1")

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
		_, err := service.SaveBatch(ctx, batchReqs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkService_DeleteURLs(b *testing.B) {
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockRepository(ctrl)
	log, _ := logger.Initialize("debug")
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "1")

	storage.EXPECT().
		DeleteURLBatch(gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	service := NewService(baseURL, storage, log)

	urls := []string{strings.Repeat("yandex.ru", 1000)}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := service.DeleteURLs(ctx, urls)
		if err != nil {
			b.Fatal(err)
		}
	}
}
