package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
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

var (
	jwtSecret    string
	adminUser    string
	adminPassEnc string // Encrypted password
)

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logging.Error("FATAL: JWT_SECRET environment variable not set. Set it before starting the server.")
		os.Exit(1)
	}
	if len(jwtSecret) < 32 {
		logging.Warn("JWT_SECRET is less than 32 characters. Recommended length is 32+ characters for security.")
	}

	// Load admin credentials from environment
	adminUser = os.Getenv("ADMIN_USER")
	adminPass := os.Getenv("ADMIN_PASS")

	if adminUser == "" || adminPass == "" {
		logging.Error("FATAL: ADMIN_USER and ADMIN_PASS environment variables not set. Set them before starting the server.")
		os.Exit(1)
	}

	// Encrypt the password using JWT_SECRET as the encryption key
	var err error
	adminPassEnc, err = encryptPassword(adminPass, jwtSecret)
	if err != nil {
		logging.Errorf("FATAL: Failed to encrypt admin password: %v", err)
		os.Exit(1)
	}

	logging.Infof("admin user configured: %s", adminUser)
}

// encryptPassword encrypts a plaintext password using AES-256-GCM with the JWT_SECRET as key
func encryptPassword(plaintext, key string) (string, error) {
	// Derive a 32-byte key from JWT_SECRET using SHA256
	hash := sha256.Sum256([]byte(key))
	encryptionKey := hash[:]

	// Create cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "failed to create GCM")
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "failed to generate nonce")
	}

	// Encrypt and return as hex string (nonce + ciphertext)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts a password encrypted with encryptPassword
func decryptPassword(encrypted, key string) (string, error) {
	// Derive the same 32-byte key from JWT_SECRET using SHA256
	hash := sha256.Sum256([]byte(key))
	decryptionKey := hash[:]

	// Decode hex
	ciphertext, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode encrypted password")
	}

	// Create cipher
	block, err := aes.NewCipher(decryptionKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "failed to create GCM")
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to decrypt password")
	}

	return string(plaintext), nil
}

// ValidateAdminLogin checks if the provided username and password are valid
// Note: Password is encrypted with JWT_SECRET and compared via decryption
func ValidateAdminLogin(username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, errors.New("username and password are required")
	}

	// Check if username matches
	if username != adminUser {
		logging.Warnf("login attempt with invalid username: %s", username)
		return false, nil
	}

	// Decrypt stored password and compare
	decrypted, err := decryptPassword(adminPassEnc, jwtSecret)
	if err != nil {
		logging.Errorf("failed to decrypt admin password: %v", err)
		return false, err
	}

	// Simple string comparison (password is only encrypted in storage)
	if password != decrypted {
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
