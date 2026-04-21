package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/username/app-recrutamento-ia/docs" // Swagger docs
	"github.com/username/app-recrutamento-ia/internal/handlers"
	"github.com/username/app-recrutamento-ia/internal/server/middleware"
)

// NewRouter creates and configures the main HTTP router for the application.
func NewRouter(sessionHandler *handlers.SessionHandler) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger)
	r.Use(middleware.MetricsMiddleware) // Prometheus Metrics
	r.Use(middleware.SecurityHeaders)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Healthcheck endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Prometheus metrics endpoint
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Setup API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "active"}`))
		})

		// Session routes
		if sessionHandler != nil {
			r.Route("/sessions", func(r chi.Router) {
				r.Get("/{id}", sessionHandler.GetSession)
				r.Post("/{id}/start", sessionHandler.StartSession)
			})
		}
	})

	return r
}
