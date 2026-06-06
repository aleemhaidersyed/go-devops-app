package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// responseWriter wraps http.ResponseWriter to capture the status code
// because Go's default ResponseWriter doesn't let you read the status after writing
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader intercepts the status code before passing it to the real writer
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger is our middleware function
// It wraps every handler and logs request details automatically
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Record when request started

		// Wrap the response writer so we can capture the status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default to 200
		}

		// Call the actual handler
		next.ServeHTTP(wrapped, r)

		// After handler finishes, log everything
		log.Info().
			Str("method", r.Method).           // GET, POST, DELETE etc.
			Str("path", r.URL.Path).           // /health, /tasks etc.
			Int("status", wrapped.statusCode). // 200, 201, 404 etc.
			Dur("duration", time.Since(start)).// How long it took
			Str("ip", r.RemoteAddr).           // Caller's IP address
			Msg("request completed")
	})
}
