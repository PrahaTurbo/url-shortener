package service

import (
	"fmt"
	"testing"

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

func TestService_generateID(t *testing.T) {
	s := &Service{
		URLs: &storageMock{db: make(map[string][]byte)},
	}

	tests := []struct {
		name      string
		originURL []byte
		want      string
	}{
		{
			name:      "short url",
			originURL: []byte("https://yandex.ru"),
			want:      "FgAJzm",
		},
		{
			name: "long url",
			originURL: []byte("https://ya.ru/showcaptcha?cc=1&mt=556239AC7B55DDEC0C06BBA1F6D6E2985D9F23603CAA1FF1DE1570FD960655C" +
				"09742E97F4E3F557D3E0215CB02799693345D3F44BD26CDC971851D3F7C06C17AA43B3F8C793D92C6F562F3A9361005BF6BCFFA7B35DE3F4531D1" +
				"&retpath=aHR0cHM6Ly95YS5ydS8__0b4e4aaaea7aedcb402e438c986228bc&t=2/1685895519/8234027974e84ca1528e1a19f6ac645a&u=370f6" +
				"9e3-9f254057-c5a3e5d4-730886de&s=5e681ad4b522c86ec4bd1081837b33c1"),
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
		URLs: &storageMock{db: make(map[string][]byte)},
	}

	tests := []struct {
		name string
		url  []byte
		want []byte
	}{
		{
			name: "save url",
			url:  []byte("https://yandex.ru"),
			want: []byte("https://yandex.ru"),
		},
		{
			name: "don't save same url",
			url:  []byte("https://yandex.ru"),
			want: []byte("https://yandex.ru"),
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
		url []byte
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
				url: []byte("https://yandex.ru"),
				err: false,
			},
		},
		{
			name: "get url with false id",
			id:   "id2",
			want: want{
				url: []byte("https://yandex.ru"),
				err: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				URLs: &storageMock{db: map[string][]byte{"id": tt.want.url}},
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
