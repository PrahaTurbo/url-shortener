package middleware

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"
)

// Decompress middleware checks if the request body is gzip compressed.
// If it is, the middleware decompresses the request body using gzip.
func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					log.Println("failed to close gzip.Reader:", err)
				}
			}()
		}

		next.ServeHTTP(w, r)
	})
}
