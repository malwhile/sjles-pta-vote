package services

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/logging"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

var jwtSecret string

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
		logging.Warn("JWT_SECRET not set, using default value. Change this in production!")
	}
}

// GetAdminCredentials retrieves admin credentials from environment variables
// Format: ADMIN_USERS=username:password|username2:password2
func getAdminCredentials() map[string]string {
	adminUsers := os.Getenv("ADMIN_USERS")
	if adminUsers == "" {
		// Default admin user (change in production)
		adminUsers = "admin:admin"
		logging.Warn("ADMIN_USERS not set, using default admin:admin")
	}

	credentials := make(map[string]string)
	for _, userPass := range strings.Split(adminUsers, "|") {
		parts := strings.Split(strings.TrimSpace(userPass), ":")
		if len(parts) == 2 {
			credentials[parts[0]] = parts[1]
		}
	}
	return credentials
}

// hashPassword hashes a password using SHA256
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ValidateAdminLogin checks if the provided username and password are valid
func ValidateAdminLogin(username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, errors.New("username and password are required")
	}

	credentials := getAdminCredentials()
	storedPassword, exists := credentials[username]

	if !exists {
		// Return false but not an error for security reasons (don't reveal if user exists)
		return false, nil
	}

	// Compare passwords (you could enhance this with bcrypt in production)
	if storedPassword != password {
		return false, nil
	}

	return true, nil
}

// GenerateAuthToken generates a JWT token for an authenticated admin user
func GenerateAuthToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", errors.Wrap(err, "failed to generate token")
	}

	return tokenString, nil
}

// VerifyAuthToken verifies a JWT token and returns the username if valid
func VerifyAuthToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", errors.Wrap(err, "failed to parse token")
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	username, ok := (*claims)["username"].(string)
	if !ok {
		return "", errors.New("username not found in token")
	}

	return username, nil
}

// LogoutHandler handles admin logout (POST /api/admin/logout)
// Note: With JWT tokens, logout is primarily a client-side operation.
// This endpoint confirms the logout action on the server side.
func LogoutHandler(resWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the username from the token for audit logging
	authHeader := request.Header.Get("Authorization")
	if authHeader != "" {
		if username, err := VerifyAuthToken(authHeader[7:]); err == nil { // Remove "Bearer " prefix
			logging.Audit("LOGOUT", username, "user logged out", true)
		}
	}

	// Send success response
	// The client should remove the token from localStorage
	common.SendSuccess(resWriter, map[string]interface{}{
		"message": "Logged out successfully. Please clear your authentication token.",
	})
}
