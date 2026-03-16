package middleware

import (
	"net/http"
	"strings"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/logging"
	"go-sjles-pta-vote/server/services"
)

// AuthMiddleware verifies JWT authentication on protected endpoints
// Expects Authorization header in format: "Bearer <token>"
// Returns 401 Unauthorized if token is missing or invalid
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logging.Warnf("unauthorized request attempt to %s without token", r.URL.Path)
			common.SendError(w, "Missing authorization token", http.StatusUnauthorized)
			return
		}

		// Extract bearer token from "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logging.Warnf("invalid authorization header format for %s", r.URL.Path)
			common.SendError(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		username, err := services.VerifyAuthToken(token)
		if err != nil {
			logging.Warnf("invalid or expired token for request to %s: %v", r.URL.Path, err)
			common.SendError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add username to context for audit logging
		r.Header.Set("X-Username", username)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// GetUsernameFromContext extracts the authenticated username from request headers
// Returns "unknown" if not found (middleware failed to extract it)
func GetUsernameFromContext(r *http.Request) string {
	username := r.Header.Get("X-Username")
	if username == "" {
		return "unknown"
	}
	return username
}
