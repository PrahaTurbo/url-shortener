package service

import (
	"errors"
	"github.com/PrahaTurbo/url-shortener/internal/models"
	"github.com/PrahaTurbo/url-shortener/internal/storage"
	"github.com/PrahaTurbo/url-shortener/internal/storage/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

var baseURL = "localhost:8080"

func setupService(mockStorage *mock.MockRepository) Service {
	return Service{
		Storage: mockStorage,
		baseURL: baseURL,
	}
}

func TestService_generateID(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	srv := setupService(s)

	tests := []struct {
		name      string
		originURL string
		want      string
	}{
		{
			name:      "short url",
			originURL: "https://yandex.ru",
			want:      "FgAJzm",
		},
		{
			name: "long url",
			originURL: "https://ya.ru/showcaptcha?cc=1&mt=556239AC7B55DDEC0C06BBA1F6D6E2985D9F23603CAA1FF1DE1570FD960655C" +
				"09742E97F4E3F557D3E0215CB02799693345D3F44BD26CDC971851D3F7C06C17AA43B3F8C793D92C6F562F3A9361005BF6BCFFA7B35DE3F4531D1" +
				"&retpath=aHR0cHM6Ly95YS5ydS8__0b4e4aaaea7aedcb402e438c986228bc&t=2/1685895519/8234027974e84ca1528e1a19f6ac645a&u=370f6" +
				"9e3-9f254057-c5a3e5d4-730886de&s=5e681ad4b522c86ec4bd1081837b33c1",
			want: "yp58Qz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := srv.generateID(tt.originURL)
			assert.Equal(t, tt.want, id)
		})
	}
}

func TestService_SaveURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	urlRecord := storage.URLRecord{
		UUID:        "86d0f933-287c-4e1a-9978-4d9706e3e94f",
		ShortURL:    "fpCk-c",
		OriginalURL: "https://ya.ru",
	}

	s.EXPECT().
		PutURL(gomock.Any()).
		Return(nil)

	s.EXPECT().
		GetURL(urlRecord.ShortURL).
		Return(&urlRecord, nil)

	s.EXPECT().
		GetURL("FgAJzm").
		Return(nil, errors.New("no url"))

	srv := setupService(s)

	tests := []struct {
		name    string
		url     string
		want    string
		wantErr error
	}{
		{
			name:    "save url successfully",
			url:     "https://yandex.ru",
			want:    baseURL + "/" + "FgAJzm",
			wantErr: nil,
		},
		{
			name:    "don't save url that already in storage",
			url:     urlRecord.OriginalURL,
			want:    baseURL + "/" + urlRecord.ShortURL,
			wantErr: ErrAlready,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL, err := srv.SaveURL(tt.url)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			}

			assert.Equal(t, tt.want, shortURL)
		})
	}
}

func TestService_GetURL(t *testing.T) {
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
		GetURL("abc").
		Return(nil, errors.New("no url"))

	srv := setupService(s)

	type want struct {
		url string
		err bool
	}

	tests := []struct {
		name     string
		shortURL string
		want     want
	}{
		{
			name:     "get url",
			shortURL: urlRecord.ShortURL,
			want: want{
				url: urlRecord.OriginalURL,
				err: false,
			},
		},
		{
			name:     "err when getting url with false id",
			shortURL: "abc",
			want: want{
				url: "",
				err: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			url, err := srv.GetURL(tt.shortURL)
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
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	s.EXPECT().
		GetURL(gomock.Any()).
		Return(nil, errors.New("no url")).AnyTimes()

	srv := setupService(s)

	tests := []struct {
		name    string
		batch   []models.BatchRequest
		want    []models.BatchResponse
		wantErr bool
	}{
		{
			name: "save batch successfully",
			batch: []models.BatchRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://ya.ru",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://yandex.ru",
				},
			},
			want: []models.BatchResponse{
				{
					CorrelationID: "1",
					ShortURL:      baseURL + "/fpCk-c",
				},
				{
					CorrelationID: "2",
					ShortURL:      baseURL + "/FgAJzm",
				},
			},
			wantErr: false,
		},
		{
			name: "save batch failed",
			batch: []models.BatchRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://ya.ru",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://yandex.ru",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				s.EXPECT().
					PutBatchURLs(gomock.Any()).
					Return(errors.New("cannot save batch urls"))
			} else {
				s.EXPECT().
					PutBatchURLs(gomock.Any()).
					Return(nil)
			}

			resp, err := srv.SaveBatch(tt.batch)

			if tt.wantErr {
				assert.NotEmpty(t, err)
				return
			}

			if assert.NoError(t, err) {
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestService_formURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mock.NewMockRepository(ctrl)

	srv := setupService(s)

	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "successfully form short url",
			id:   "dfdvFd",
			want: baseURL + "/" + "dfdvFd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, srv.formURL(tt.id))
		})
	}
}
