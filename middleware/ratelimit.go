package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	timestamps []time.Time
}

// rateLimiter is a simple in-memory, per-IP sliding-window limiter.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	// Periodic cleanup of stale entries
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, v := range rl.visitors {
			filtered := v.timestamps[:0]
			for _, t := range v.timestamps {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(rl.visitors, ip)
			} else {
				v.timestamps = filtered
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{}
		rl.visitors[ip] = v
	}

	// Drop timestamps outside the window
	filtered := v.timestamps[:0]
	for _, t := range v.timestamps {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	v.timestamps = filtered

	if len(v.timestamps) >= rl.limit {
		return false
	}

	v.timestamps = append(v.timestamps, now)
	return true
}

// RateLimit returns a Gin middleware that limits each client IP to `limit`
// requests per `window`. Intended for sensitive endpoints (auth).
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
