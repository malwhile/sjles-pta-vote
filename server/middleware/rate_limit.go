package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/logging"
)

// RateLimiter uses token bucket algorithm to limit requests per IP
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	maxReq   int           // max requests per window
	window   time.Duration // time window
	cleanup  time.Duration // cleanup old entries
}

// Visitor tracks request count for an IP
type Visitor struct {
	limiter *time.Ticker
	lastSeen time.Time
	count    int
}

// NewRateLimiter creates a new rate limiter
// maxReq: maximum requests per window
// window: time window (e.g., 1 minute)
// cleanup: how often to clean old entries
func NewRateLimiter(maxReq int, window, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		maxReq:   maxReq,
		window:   window,
		cleanup:  cleanup,
	}

	// Cleanup old entries periodically
	go func() {
		ticker := time.NewTicker(cleanup)
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, visitor := range rl.visitors {
				if now.Sub(visitor.lastSeen) > window {
					visitor.limiter.Stop()
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

// IsAllowed checks if the request should be allowed
func (rl *RateLimiter) IsAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		// New visitor
		ticker := time.NewTicker(rl.window)
		visitor = &Visitor{
			limiter:  ticker,
			lastSeen: time.Now(),
			count:    1,
		}
		rl.visitors[ip] = visitor
		return true
	}

	// Update last seen time
	visitor.lastSeen = time.Now()

	// Check if allowed
	if visitor.count < rl.maxReq {
		visitor.count++
		return true
	}

	return false
}

// Middleware returns a rate limiting middleware handler
// Limits: 100 votes/hour, 60 poll views/minute per IP
func RateLimitVotes(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !limiter.IsAllowed(ip) {
				logging.Warnf("rate limit exceeded for IP: %s", ip)
				common.SendError(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts client IP from request
// Checks X-Forwarded-For for proxy setups
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		ips := xff
		if idx := len(ips) - 1; idx >= 0 {
			return ips[:idx+1]
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Global rate limiters
var (
	// Vote endpoint: 100 votes per hour per IP
	VoteRateLimiter = NewRateLimiter(100, time.Hour, 10*time.Minute)

	// Poll view endpoint: 300 views per minute per IP (for high traffic)
	PollViewRateLimiter = NewRateLimiter(300, time.Minute, 5*time.Minute)
)
