package middleware

import (
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testJWTsecret = "test-secret"

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

			tokenString := got.Value
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(testJWTsecret), nil
			})

			token, err = jwt.ParseWithClaims(got.Value, &Claims{}, func(t *jwt.Token) (interface{}, error) {
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
