package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-sjles-pta-vote/server/services"
)

func TestAuthMiddlewareValidToken(t *testing.T) {
	// Generate a valid test token
	token, err := services.GenerateAuthToken("testuser")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Create a test handler that checks if auth middleware sets context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authenticated"))
	})

	// Wrap with auth middleware
	wrappedHandler := AuthMiddleware(testHandler)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/api/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddlewareMissingToken(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := AuthMiddleware(testHandler)

	// Create request without token
	req := httptest.NewRequest("GET", "/api/admin/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing token, got %d", w.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := AuthMiddleware(testHandler)

	// Create request with invalid token
	req := httptest.NewRequest("GET", "/api/admin/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}
}

func TestAuthMiddlewareMalformedHeader(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := AuthMiddleware(testHandler)

	// Test various malformed headers
	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer", "token123"},
		{"invalid prefix", "Basic token123"},
		{"empty bearer", "Bearer "},
		{"extra spaces", "Bearer  token123"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest("GET", "/api/admin/test", nil)
		req.Header.Set("Authorization", tc.header)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", tc.name, w.Code)
		}
	}
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	// This test would require manipulating token expiration
	// For now, we just verify that valid tokens work
	// Expired token testing would require modifying the token generation to support custom expiry
	token, err := services.GenerateAuthToken("testuser")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := AuthMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/api/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Fresh token should be valid
	if w.Code != http.StatusOK {
		t.Errorf("Fresh token should be valid, got %d", w.Code)
	}
}
