package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecompress(t *testing.T) {
	originalContent := "test data"

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(originalContent))
	if err != nil {
		t.Error(err)
	}

	if err := gz.Close(); err != nil {
		t.Error(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
			return
		}

		assert.Equal(t, originalContent, string(body))
	})

	req := httptest.NewRequest("GET", "http://example.com/foo", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	wrw := httptest.NewRecorder()
	middlewareHandler := Decompress(handler)
	middlewareHandler.ServeHTTP(wrw, req)
}
