package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserIDKeyType is a custom string type that will be used as a key in request context.
type UserIDKeyType string

const (
	bearerSchema   = "bearer"
	authentication = "authorization"

	// UserIDKey is the key used in context to retrieve the UserID.
	UserIDKey          UserIDKeyType = "userID"
	jwtTokenCookieName string        = "token"
)

// Claims represents the custom claims we're using in JWT tokens.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type Auth struct {
	secret        string
	trustedSubnet string
}

func NewAuth(secret string, subnet string) *Auth {
	return &Auth{
		secret:        secret,
		trustedSubnet: subnet,
	}
}

func (a *Auth) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader, ok := md[authentication]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	splits := strings.SplitN(authHeader[0], " ", 2)
	if len(splits) < 2 {
		return nil, status.Errorf(codes.Unauthenticated, "the token is not in the correct format")
	}

	if strings.ToLower(splits[0]) != bearerSchema {
		return nil, status.Errorf(codes.Unauthenticated, "the token is not a Bearer token")
	}

	tokenString := splits[1]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(a.secret), nil
	})

	if err != nil || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "the token is invalid")
	}

	if claims.UserID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user_id is empty")
	}

	newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)

	return handler(newCtx, req)
}

// BasicMiddlewareHTTP is a middleware for JWT authentication.
// The middleware checks if JWT Token is present in a cookie and valid.
// If the cookie is not present or JWT Token is not valid, a new cookie with JWT Token is created.
//
// The middleware sets the user id (from the JWT claims) in the request context.
func (a *Auth) BasicMiddlewareHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(jwtTokenCookieName)
		if err != nil {
			cookie, err = createJWTAuthCookie(a.secret)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, cookie)
		}

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(a.secret), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			newCookie, err := createJWTAuthCookie(a.secret)
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

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddlewareHTTP is a middleware function that checks if the incoming request is from a trusted subnet.
func (a *Auth) AdminMiddlewareHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.trustedSubnet == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, trustNet, err := net.ParseCIDR(a.trustedSubnet)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ipStr := r.Header.Get("X-Real-IP")
		ip := net.ParseIP(ipStr)

		if !trustNet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
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
		Name:     jwtTokenCookieName,
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
	}

	return cookie, nil
}
