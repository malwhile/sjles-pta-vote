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

var jwtSecret string

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logging.Error("FATAL: JWT_SECRET environment variable not set. Set it before starting the server.")
		os.Exit(1)
	}
	if len(jwtSecret) < 32 {
		logging.Warn("JWT_SECRET is less than 32 characters. Recommended length is 32+ characters for security.")
	}
}

// GetAdminCredentials retrieves admin credentials from environment variables
// Format: ADMIN_USERS=username:bcrypt_hash|username2:bcrypt_hash
// Where bcrypt_hash is a bcrypt-hashed password created with bcrypt.GenerateFromPassword
func getAdminCredentials() map[string]string {
	adminUsers := os.Getenv("ADMIN_USERS")
	if adminUsers == "" {
		logging.Error("FATAL: ADMIN_USERS environment variable not set. Set it before starting the server.")
		os.Exit(1)
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

// ValidateAdminLogin checks if the provided username and password are valid
// Passwords are compared using bcrypt constant-time comparison
func ValidateAdminLogin(username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, errors.New("username and password are required")
	}

	credentials := getAdminCredentials()
	storedHash, exists := credentials[username]

	if !exists {
		// Always perform a bcrypt comparison to avoid timing attacks
		// This uses a dummy hash to maintain consistent timing
		bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyhashtoavoidtimingattacks"), []byte(password))
		return false, nil
	}

	// Use bcrypt for constant-time password comparison
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
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

// HashPassword hashes a plaintext password using bcrypt
// Cost is set to 12 for strong security (OWASP recommended minimum)
// This function is exported for use in setup utilities
func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return "", errors.Wrap(err, "failed to hash password")
	}
	return string(hash), nil
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
