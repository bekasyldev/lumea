package ratelimit

import (
	"math"
	"net"
	"net/http"
	"strconv"
)

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := clientIP(r)

		if !rl.Allow(clientID) {
			// rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.FormatFloat(rl.capacity, 'f', 0, 64))
			w.Header().Set("X-Rate-Remaining", "0")
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(1/rl.tokenPerSecond))))

			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
