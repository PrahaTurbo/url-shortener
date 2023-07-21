package middleware

import (
	"context"
	"fmt"
	"github.com/PrahaTurbo/url-shortener/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func Auth(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var cookie *http.Cookie

			cookie, err := r.Cookie(string(config.JWTTokenCookieName))
			if err != nil {
				newCookie, err := createJWTAuthCookie(jwtSecret)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie = newCookie
				http.SetCookie(w, cookie)
			}

			claims := &Claims{}

			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}

				return []byte(jwtSecret), nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				newCookie, err := createJWTAuthCookie(jwtSecret)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie = newCookie
				http.SetCookie(w, cookie)
			}

			if claims.UserID == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), string(config.UserIDKey), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func createJWTAuthCookie(jwtSecret string) (*http.Cookie, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: uuid.New().String(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     string(config.JWTTokenCookieName),
		Value:    tokenString,
		HttpOnly: true,
	}

	return cookie, nil
}
