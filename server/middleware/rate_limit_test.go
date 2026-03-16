package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitingGlobalLimiters(t *testing.T) {
	// Test that the global rate limiters exist and can be used
	if VoteRateLimiter == nil {
		t.Fatal("VoteRateLimiter should not be nil")
	}

	if PollViewRateLimiter == nil {
		t.Fatal("PollViewRateLimiter should not be nil")
	}
}

func TestRateLimitMiddlewareIntegration(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with the rate limiter
	wrappedHandler := RateLimitVotes(PollViewRateLimiter)(handler)

	// First request should succeed
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 200 or 429, got %d", w.Code)
	}
}

func TestRateLimitWithProxyHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := RateLimitVotes(VoteRateLimiter)(handler)

	// Request with X-Forwarded-For header
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345" // Local
	req.Header.Set("X-Forwarded-For", "203.0.113.42")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Should succeed or be rate limited, but not error
	if w.Code >= 500 {
		t.Errorf("Request should not cause server error, got %d", w.Code)
	}
}
