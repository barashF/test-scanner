package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/metrics"
)

func MetricsMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(startTime)

			rctx := chi.RouteContext(r.Context())
			routePattern := r.URL.Path
			if rctx != nil && rctx.RoutePattern() != "" {
				routePattern = rctx.RoutePattern()
			}
			metrics.HTTPRequestDuration.WithLabelValues(r.Method, routePattern, strconv.Itoa(rw.statusCode)).Observe(duration.Seconds())
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
