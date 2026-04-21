package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal counts the total number of HTTP requests processed
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"method", "path", "status"},
	)

	// HttpRequestDuration measures the duration of HTTP requests
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// InterviewSessionsActive counts currently active interview sessions
	InterviewSessionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "interview_sessions_active",
			Help: "Current number of active interview sessions.",
		},
	)
)
