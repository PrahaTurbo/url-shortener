package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminAuth(t *testing.T) {
	trustedSubnet := "192.0.2.0/24"
	untrustedIP := "198.51.100.1" // not in trustedSubnet
	trustedIP := "192.0.2.1"      // in trustedSubnet

	tests := []struct {
		name           string
		subnet         string
		requestIP      string
		expectedStatus int
	}{
		{
			name:           "Request From Trusted IP",
			subnet:         trustedSubnet,
			requestIP:      trustedIP,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Request From Untrusted IP",
			subnet:         trustedSubnet,
			requestIP:      untrustedIP,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid Subnet",
			subnet:         "invalid subnet",
			requestIP:      trustedIP,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Empty Subnet",
			subnet:         "",
			requestIP:      trustedIP,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Real-IP", tt.requestIP)

			rr := httptest.NewRecorder()

			handler := AdminAuth(tt.subnet)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
