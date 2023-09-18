package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_generateShortURL(t *testing.T) {
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
			id := generateShortURL(tt.originURL)
			assert.Equal(t, tt.want, id)
		})
	}
}

func BenchmarkService_generateShortURL(b *testing.B) {
	originalURL := "https://ya.ru/showcaptcha?cc=1&mt=556239AC7B55DDEC0C06BBA1F6D6E2985D9F23603CAA1FF1DE1570FD960655C" +
		"09742E97F4E3F557D3E0215CB02799693345D3F44BD26CDC971851D3F7C06C17AA43B3F8C793D92C6F562F3A9361005BF6BCFFA7B35DE3F4531D1" +
		"&retpath=aHR0cHM6Ly95YS5ydS8__0b4e4aaaea7aedcb402e438c986228bc&t=2/1685895519/8234027974e84ca1528e1a19f6ac645a&u=370f6" +
		"9e3-9f254057-c5a3e5d4-730886de&s=5e681ad4b522c86ec4bd1081837b33c1"

	for i := 0; i < b.N; i++ {
		generateShortURL(originalURL)
	}
}
