package middleware

import (
	"net"
	"net/http"
)

// AdminAuth is a middleware function that checks if the incoming request is from a trusted subnet.
func AdminAuth(subnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, trustNet, err := net.ParseCIDR(subnet)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)

			if !trustNet.Contains(ip) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
