package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity       float64
	tokens         float64
	tokenPerSecond float64
	lastRefill     time.Time
	mu             sync.Mutex // protects from concurrent access
}

func NewTokenBucket(capacity float64, tokenPerSecond float64) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity, // start with full token available
		tokenPerSecond: tokenPerSecond,
		lastRefill:     time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tokenToAdd := elapsed * tb.tokenPerSecond
	tb.tokens = min(tb.capacity, tb.tokens+tokenToAdd) // should not exceed capacity
	tb.lastRefill = now
}

func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

func (tb *TokenBucket) AllowN(n float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

func (tb *TokenBucket) TokenAvailable() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}
