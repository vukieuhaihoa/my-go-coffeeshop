package gateway

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// LogRequests is a middleware that logs incoming HTTP requests and their response status.
// It wraps the provided http.Handler and logs the request method, path, status code, and duration.
// Parameters:
//   - log: The logger to be used for logging requests.
//   - next: The next http.Handler to be called in the middleware chain.
//
// Returns:
//   - An http.Handler that logs requests and calls the next handler.
func LogRequests(log logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"took", time.Since(start))
	})
}

// AllowCORS is a middleware that allows Cross-Origin Resource Sharing (CORS) for specified origins.
// It wraps the provided http.Handler and adds the necessary CORS headers to the response.
// Parameters:
//   - allowed: A slice of allowed origins for CORS.
//   - next: The next http.Handler to be called in the middleware chain.
//
// Returns:
//   - An http.Handler that allows CORS for specified origins and calls the next handler.
func AllowCORS(allowed []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		origin := r.Header.Get("Origin")
		if origin == "" || !slices.Contains(allowed, origin) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)

		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// WithTimeout is a middleware that sets a timeout for HTTP requests.
// It wraps the provided http.Handler and cancels the request context if it exceeds the specified duration.
// Parameters:
//   - d: The duration for the timeout.
//   - next: The next http.Handler to be called in the middleware chain.
//
// Returns:
//   - An http.Handler that sets a timeout for requests and calls the next handler.
func WithTimeout(d time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
