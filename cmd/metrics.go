package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var (
	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration",
		Help: "Duration of HTTP request.",
		Buckets: []float64{
			0.001, // 1ms
			0.005, // 5ms
			0.01,  // 10ms
			0.025, // 25ms
			0.05,  // 50ms
			0.1,   // 100ms
			0.25,  // 250ms
			0.5,   // 500ms
			1.0,   // 1s
			2.5,   // 2.5s
			5.0,   // 5s
			10.0,  // 10s
		},
	}, []string{"pattern"})
	httpRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_count",
		Help: "Total number of HTTP requests received",
	}, []string{"route", "method", "status"})
)

func (app *application) metricMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		requestPattern := r.Pattern
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues(requestPattern).Observe(duration)
		httpRequestTotal.WithLabelValues(requestPattern, r.Method, strconv.Itoa(rw.statusCode)).Inc()
	})
}
