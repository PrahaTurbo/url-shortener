package service

import (
	"github.com/PrahaTurbo/url-shortener/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_generateID(t *testing.T) {
	s := &Service{
		URLs: &mock.StorageMock{DB: make(map[string]string)},
	}

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
			id := s.generateID(tt.originURL)
			assert.Equal(t, tt.want, id)
		})
	}
}

func TestService_SaveURL(t *testing.T) {
	s := &Service{
		URLs: &mock.StorageMock{DB: make(map[string]string)},
	}

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "save url",
			url:  "https://yandex.ru",
			want: "https://yandex.ru",
		},
		{
			name: "don't save same url",
			url:  "https://yandex.ru",
			want: "https://yandex.ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := s.SaveURL(tt.url)
			url, err := s.GetURL(id)

			if assert.NoError(t, err) {
				assert.Equal(t, tt.want, url)
			}
		})
	}
}

func TestService_GetURL(t *testing.T) {
	type want struct {
		url string
		err bool
	}

	tests := []struct {
		name string
		id   string
		want want
	}{
		{
			name: "get url",
			id:   "id",
			want: want{
				url: "https://yandex.ru",
				err: false,
			},
		},
		{
			name: "get url with false id",
			id:   "id2",
			want: want{
				url: "https://yandex.ru",
				err: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				URLs: &mock.StorageMock{DB: map[string]string{"id": tt.want.url}},
			}

			url, err := s.GetURL(tt.id)
			if !tt.want.err {
				require.NoError(t, err)

				assert.Equal(t, url, tt.want.url)
				return
			}

			assert.Error(t, err)
		})
	}
}
