package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testJWTsecret = "test-secret"

func TestAuth_AdminMiddlewareHTTP(t *testing.T) {
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

			auth := NewAuth(testJWTsecret, tt.subnet)

			handler := auth.AdminMiddlewareHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestAuth_BasicMiddlewareHTTP(t *testing.T) {
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

			auth := NewAuth(testJWTsecret, "")

			auth.BasicMiddlewareHTTP(http.HandlerFunc(handler)).ServeHTTP(rr, req)
			assert.Equal(t, test.status, rr.Code)

			if test.status == http.StatusOK {
				assert.NotNil(t, userID, "userID should not be nil")
			}
		})
	}
}

func Test_createJWTAuthCookie(t *testing.T) {
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

type mockHandler func(ctx context.Context, req interface{}) (interface{}, error)

func (m mockHandler) Handle(ctx context.Context, req interface{}) (interface{}, error) {
	return m(ctx, req)
}

func TestAuth_UnaryServerInterceptor(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: uuid.New().String(),
	})

	tokenString, _ := token.SignedString([]byte(testJWTsecret))

	testCases := []struct {
		name        string
		ctx         context.Context
		expectedErr codes.Code
	}{
		{
			name: "successful request",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": fmt.Sprintf("bearer %s", tokenString),
			})),
			expectedErr: codes.OK,
		},
		{
			name:        "missing metadata",
			ctx:         context.Background(),
			expectedErr: codes.Unauthenticated,
		},
		{
			name: "no authorization header",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"auth": fmt.Sprintf("bearer %s", tokenString),
			})),
			expectedErr: codes.Unauthenticated,
		},
		{
			name: "the token is not in the correct format",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": fmt.Sprintf("bearer%s", tokenString),
			})),
			expectedErr: codes.Unauthenticated,
		},
		{
			name: "the token is not a Bearer token",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": fmt.Sprintf("scheme %s", tokenString),
			})),
			expectedErr: codes.Unauthenticated,
		},
		{
			name: "invalid token",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"authorization": fmt.Sprintf("bearer %s", tokenString[1:]),
			})),
			expectedErr: codes.Unauthenticated,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			handler := mockHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			})

			a := NewAuth(testJWTsecret, "")

			_, err := a.UnaryServerInterceptor(tt.ctx, "request", nil, handler.Handle)

			if status.Code(err) != codes.OK {
				fmt.Println(err.Error())
				assert.Equal(t, tt.expectedErr.String(), status.Code(err).String())
			}
		})
	}
}
