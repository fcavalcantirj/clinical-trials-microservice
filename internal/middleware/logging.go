package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// responseWriter wraps http.ResponseWriter to capture status code and body size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bodySize   int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	written, err := rw.ResponseWriter.Write(b)
	rw.bodySize += written
	return written, err
}

// RequestIDKey is the key used to store request ID in context
type RequestIDKey struct{}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Skip logging for health check endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Generate request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context for downstream handlers
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDKey{}, requestID)
		r = r.WithContext(ctx)

		// Create logger with request context
		logger := log.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Str("ip", getClientIP(r)).
			Str("user_agent", r.UserAgent()).
			Logger()

		// Wrap response writer to capture status and size
		rw := newResponseWriter(w)

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request
		event := logger.Info().
			Int("status", rw.statusCode).
			Int64("duration_ms", duration.Milliseconds()).
			Int("body_size", rw.bodySize)

		// Add error context for 4xx and 5xx responses
		if rw.statusCode >= 400 {
			event = logger.Error().
				Int("status", rw.statusCode).
				Int64("duration_ms", duration.Milliseconds()).
				Int("body_size", rw.bodySize)
		}

		event.Msg("Request completed")
	})
}

// RequestIDMiddleware adds request ID to context and response headers
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDKey{}, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
