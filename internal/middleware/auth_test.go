package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testJWTsecret = "test-secret"

func TestAuth(t *testing.T) {
	tests := []struct {
		name   string
		cookie *http.Cookie
		status int
	}{
		{
			name:   "invalid JWT",
			cookie: &http.Cookie{Name: jwtTokenCookieName, Value: "invalidJwtToken"},
			status: http.StatusUnauthorized,
		},
		{
			name: "valid JWT",
			cookie: func() *http.Cookie {
				c, _ := createJWTAuthCookie(testJWTsecret)

				return c
			}(),
			status: http.StatusOK,
		},
		{
			name:   "invalid cookie name",
			cookie: &http.Cookie{Name: "invalid name"},
			status: http.StatusOK,
		},
		{
			name:   "user id is empty in claims",
			cookie: &http.Cookie{Name: jwtTokenCookieName, Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiIifQ.PaDvfASoSWvIy1N_mk6kMWua8kMe27k7pfSxe7vK17I"},
			status: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.AddCookie(test.cookie)
			rr := httptest.NewRecorder()

			var userID interface{}
			handler := func(w http.ResponseWriter, r *http.Request) {
				userID = r.Context().Value(UserIDKey)
			}
			authMiddleware := Auth(testJWTsecret)

			authMiddleware(http.HandlerFunc(handler)).ServeHTTP(rr, req)
			assert.Equal(t, test.status, rr.Code)

			if test.status == http.StatusOK {
				assert.NotNil(t, userID, "userID should not be nil")
			}
		})
	}
}

func TestCreateJWTAuthCookie(t *testing.T) {
	type want struct {
		name     string
		httpOnly bool
		path     string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "should return cookie with valid token",
			want: want{
				name:     jwtTokenCookieName,
				httpOnly: true,
				path:     "/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createJWTAuthCookie(testJWTsecret)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.name, got.Name)
			assert.Equal(t, tt.want.httpOnly, got.HttpOnly)
			assert.Equal(t, tt.want.path, got.Path)

			token, err := jwt.ParseWithClaims(got.Value, &Claims{}, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}

				return []byte(testJWTsecret), nil
			})

			assert.NoError(t, err)
			assert.True(t, token.Valid)
		})
	}
}
