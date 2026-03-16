package middleware

import (
	"os"
)

func init() {
	// Set up test environment variables before any other init() functions
	// This file is loaded first due to underscore prefix in filename sorting
	os.Setenv("JWT_SECRET", "test-secret-key-12345")
	os.Setenv("ADMIN_USER", "testadmin")
	os.Setenv("ADMIN_PASS", "testpass")
}
