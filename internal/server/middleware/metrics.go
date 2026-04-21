package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/username/app-recrutamento-ia/internal/observability"
)

// MetricsMiddleware logs the duration and status code of HTTP requests to Prometheus
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(ww.Status())

		// Avoid cardinality explosion by grouping paths (e.g. /api/v1/users/123 -> /api/v1/users/{id})
		// For MVP, we use the raw path. In production, use chi.RouteContext(r.Context()).RoutePattern()
		path := r.URL.Path

		observability.HttpRequestsTotal.WithLabelValues(r.Method, path, statusStr).Inc()
		observability.HttpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}
