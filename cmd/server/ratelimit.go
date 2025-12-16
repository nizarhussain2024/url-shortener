package main

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

var urlRateLimiter = &RateLimiter{
	requests: make(map[string][]time.Time),
	limit:    100, // 100 requests per window
	window:   time.Minute,
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	if times, exists := rl.requests[ip]; exists {
		validTimes := []time.Time{}
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[ip] = validTimes
	}

	if len(rl.requests[ip]) >= rl.limit {
		return false
	}

	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !urlRateLimiter.Allow(ip) {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}


