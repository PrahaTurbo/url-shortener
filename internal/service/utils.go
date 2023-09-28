package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/PrahaTurbo/url-shortener/internal/middleware"
)

// Error variables, used in the service
var (
	ErrExtractFromContext = errors.New("cannot extract userID from context")
	ErrAlready            = errors.New("URL already in storage")
	ErrNoOriginalURL      = errors.New("no url in original_url field")
)

func extractUserIDFromCtx(ctx context.Context) (string, error) {
	userIDVal := ctx.Value(middleware.UserIDKey)
	userID, ok := userIDVal.(string)
	if !ok {
		return "", fmt.Errorf("cannot extract userID from context")
	}

	return userID, nil
}

func formURL(baseURL, shortURL string) string {
	return baseURL + "/" + shortURL
}

func generateShortURL(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	truncatedHash := hash[:6]

	return truncatedHash
}
