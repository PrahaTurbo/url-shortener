package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Logger is a struct that embeds zap.Logger, bringing structured, leveled logging.
type Logger struct {
	*zap.Logger
}

// Write writes data to the http response, and records the size of data written.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader sends an HTTP response header with the provided status code, and records this status.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Initialize initializes a new Logger with the provided AtomicLevel string for level.
func Initialize(level string) (*Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

// RequestLogger is a middleware to log the message about request handling.
// This includes: the URI of the request, the HTTP method, the duration of handling, the status and size of the response.
func (logger *Logger) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Info("HTTP request handled",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("response status", responseData.status),
			zap.Int("response size", responseData.size),
		)
	})
}
