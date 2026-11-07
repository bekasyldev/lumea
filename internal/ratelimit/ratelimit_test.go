package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTooManyRequests(t *testing.T) {
	limiter := NewRateLimiter(10, 1)
	defer limiter.Stop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := limiter.Middleware(next)

	for i := 1; i <= 11; i++ {
		req := httptest.NewRequest("GET", "http://testing", nil) // same RemoteAddr => same client
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		want := http.StatusOK
		if i == 11 {
			want = http.StatusTooManyRequests
		}
		if rec.Code != want {
			t.Fatalf("request %d: got status %d, want %d", i, rec.Code, want)
		}
		if i == 11 && rec.Header().Get("Retry-After") == "" {
			t.Error("request 11: missing Retry-After header")
		}
	}
}
