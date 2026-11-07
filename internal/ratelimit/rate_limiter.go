package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter struct {
	buckets        map[string]*TokenBucket
	capacity       float64
	tokenPerSecond float64
	mu             sync.Mutex

	// cleanup settings
	cleanupInterval time.Duration
	maxIdleTime     time.Duration
	stopCleanup     chan struct{}
}

func NewRateLimiter(capacity, tokenPerSecond float64) *RateLimiter {
	rl := &RateLimiter{
		buckets:         make(map[string]*TokenBucket),
		capacity:        capacity,
		tokenPerSecond:  tokenPerSecond,
		cleanupInterval: time.Minute,
		maxIdleTime:     time.Hour,
		stopCleanup:     make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) Allow(clientId string) bool {
	return rl.AllowN(1, clientId)
}

func (rl *RateLimiter) AllowN(n float64, clientId string) bool {
	bucket := rl.getBucket(clientId)
	return bucket.AllowN(n)
}

func (rl *RateLimiter) getBucket(clientId string) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket, exists := rl.buckets[clientId]

	if exists {
		return bucket
	}

	bucket = NewTokenBucket(rl.capacity, rl.tokenPerSecond)
	rl.buckets[clientId] = bucket
	return bucket
}

// remove idle buckets to prevent memory leaks
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval) // channel that ticks after duration
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup: // to shut cleanup
			return
		}

	}
}

// remove buckets that haven't been used recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-rl.maxIdleTime)
	for clientId, bucket := range rl.buckets {
		bucket.mu.Lock()
		defer bucket.mu.Unlock()

		if bucket.lastRefill.Before(threshold) {
			delete(rl.buckets, clientId)
		}
	}
}

// stop shut down the cleanup goroutines
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}
