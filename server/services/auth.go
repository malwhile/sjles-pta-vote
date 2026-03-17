package services

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

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

var (
	jwtSecret      string
	adminUser      string
	adminPassHash  string // Bcrypt password hash
)

// isRunningTests returns true if the application is running under test
func isRunningTests() bool {
	for _, arg := range os.Args {
		if strings.Contains(arg, "test") {
			return true
		}
	}
	return false
}

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Use a default for tests, exit for production
		if isRunningTests() {
			jwtSecret = "test-secret-key-12345"
		} else {
			logging.Error("FATAL: JWT_SECRET environment variable not set. Set it before starting the server.")
			os.Exit(1)
		}
	}
	if len(jwtSecret) < 32 {
		logging.Warn("JWT_SECRET is less than 32 characters. Recommended length is 32+ characters for security.")
	}

	// Load admin credentials from environment
	adminUser = os.Getenv("ADMIN_USER")
	adminPass := os.Getenv("ADMIN_PASS")

	if adminUser == "" || adminPass == "" {
		// Use defaults for tests, exit for production
		if isRunningTests() {
			adminUser = "testadmin"
			adminPass = "testpass"
		} else {
			logging.Error("FATAL: ADMIN_USER and ADMIN_PASS environment variables not set. Set them before starting the server.")
			os.Exit(1)
		}
	}

	// Hash the password using bcrypt
	var err error
	adminPassHash, err = hashPassword(adminPass)
	if err != nil {
		logging.Errorf("FATAL: Failed to hash admin password: %v", err)
		os.Exit(1)
	}

	logging.Infof("admin user configured: %s", adminUser)
}

// hashPassword hashes a plaintext password using bcrypt
func hashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "failed to hash password")
	}
	return string(hash), nil
}

// verifyPassword verifies a plaintext password against a bcrypt hash
func verifyPassword(hash, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}

// ValidateAdminLogin checks if the provided username and password are valid
// Passwords are hashed with bcrypt and compared in a timing-safe manner
func ValidateAdminLogin(username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, errors.New("username and password are required")
	}

	// Check if username matches
	if username != adminUser {
		logging.Warnf("login attempt with invalid username: %s", username)
		return false, nil
	}

	// Verify password against bcrypt hash
	err := verifyPassword(adminPassHash, password)
	if err != nil {
		logging.Warnf("login attempt with invalid password for user: %s", username)
		return false, nil
	}

	logging.Infof("successful login for admin user: %s", username)
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
