// Package http provides HTTP middleware and handlers
package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket rate limiting per client IP
type RateLimiter struct {
	clients map[string]*clientBucket
	mu      sync.RWMutex
	rate    int           // requests per window
	window  time.Duration // time window
}

type clientBucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: max requests per window
// window: time window duration (e.g., 1 minute)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientBucket),
		rate:    rate,
		window:  window,
	}

	// Cleanup expired clients every 5 minutes
	go rl.cleanupClients()

	return rl
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rl.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"message": "too many requests, please try again later",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if a request from clientIP is allowed
func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mu.RLock()
	bucket, exists := rl.clients[clientIP]
	rl.mu.RUnlock()

	if !exists {
		// Create new bucket for this client
		bucket = &clientBucket{
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}
		rl.mu.Lock()
		rl.clients[clientIP] = bucket
		rl.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens if window has passed
	now := time.Now()
	if now.Sub(bucket.lastRefill) >= rl.window {
		bucket.tokens = rl.rate
		bucket.lastRefill = now
	}

	// Check if tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// cleanupClients removes inactive clients periodically
func (rl *RateLimiter) cleanupClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, bucket := range rl.clients {
			bucket.mu.Lock()
			// Remove clients inactive for more than 10 minutes
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.clients, ip)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}










