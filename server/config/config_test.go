package config

import (
	"os"
	"testing"
)

func TestConfigLoading(t *testing.T) {
	// Reset global config for testing
	conf = nil

	// Create a temporary .env file
	envContent := `db_path="test.db"
redis_host="localhost:6379"
redis_password="secret123"
`
	tmpFile := ".env.test"
	err := os.WriteFile(tmpFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Temporarily change config path
	oldPath := conf_path
	conf_path = tmpFile
	defer func() { conf_path = oldPath }()

	config := GetConfig()

	if config.DBPath != "test.db" {
		t.Errorf("Expected db_path='test.db', got '%s'", config.DBPath)
	}
	if config.RedisHost != "localhost:6379" {
		t.Errorf("Expected redis_host='localhost:6379', got '%s'", config.RedisHost)
	}
	if config.RedisPassword != "secret123" {
		t.Errorf("Expected redis_password='secret123', got '%s'", config.RedisPassword)
	}
}

func TestConfigDefaults(t *testing.T) {
	// Reset global config for testing
	conf = nil

	// Create a minimal .env file with only db_path
	envContent := `db_path="test.db"`
	tmpFile := ".env.test"
	err := os.WriteFile(tmpFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Temporarily change config path
	oldPath := conf_path
	conf_path = tmpFile
	defer func() { conf_path = oldPath }()

	config := GetConfig()

	if config.DBPath != "test.db" {
		t.Errorf("Expected db_path='test.db', got '%s'", config.DBPath)
	}
	// Should use defaults for missing values
	if config.RedisHost != defaults["redis_host"] {
		t.Errorf("Expected redis_host default, got '%s'", config.RedisHost)
	}
}

func TestConfigQuotedValues(t *testing.T) {
	// Reset global config for testing
	conf = nil

	// Test both single and double quoted values
	envContent := `db_path="quoted_with_double.db"
redis_host='quoted_with_single.com'
`
	tmpFile := ".env.test"
	err := os.WriteFile(tmpFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Temporarily change config path
	oldPath := conf_path
	conf_path = tmpFile
	defer func() { conf_path = oldPath }()

	config := GetConfig()

	if config.DBPath != "quoted_with_double.db" {
		t.Errorf("Failed to strip double quotes, got '%s'", config.DBPath)
	}
	if config.RedisHost != "quoted_with_single.com" {
		t.Errorf("Failed to strip single quotes, got '%s'", config.RedisHost)
	}
}

func TestConfigCommentsAndBlankLines(t *testing.T) {
	// Reset global config for testing
	conf = nil

	// Test that comments and blank lines are ignored
	envContent := `# This is a comment
db_path="test.db"

# Another comment
redis_host="localhost"

redis_password="pass"
`
	tmpFile := ".env.test"
	err := os.WriteFile(tmpFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Temporarily change config path
	oldPath := conf_path
	conf_path = tmpFile
	defer func() { conf_path = oldPath }()

	config := GetConfig()

	if config.DBPath != "test.db" {
		t.Errorf("Failed to parse config with comments, got '%s'", config.DBPath)
	}
}
