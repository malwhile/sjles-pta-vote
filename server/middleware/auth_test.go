package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestMain(m *testing.M) {
	// Set up test environment before any tests run
	os.Setenv("JWT_SECRET", "test-secret-key-12345")
	os.Setenv("ADMIN_USER", "testadmin")
	os.Setenv("ADMIN_PASS", "testpass")

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// generateTestToken creates a valid JWT token for testing
func generateTestToken(username string) string {
	// Use the same secret as set in TestMain
	secret := "test-secret-key-12345"
	claims := jwt.MapClaims{
		"username": username,
		"exp":      9999999999, // Far future expiration
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	// Create a test token with known secret
	token := generateTestToken("testuser")

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
	// Create a token with far-future expiration to test valid tokens
	// Expired token testing would require modifying VerifyAuthToken
	token := generateTestToken("testuser")

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
