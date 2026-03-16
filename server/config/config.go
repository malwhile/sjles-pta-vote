package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Config struct with tags for proper mapping
type Config struct {
	DBPath        string `env:"db_path"`
	RedisHost     string `env:"redis_host"`
	RedisPassword string `env:"redis_password"`
}

var conf *Config
var conf_path string = ".env"

// Default configuration values
var defaults = map[string]string{
	"db_path":        "./sjles-pta-vote.db",
	"redis_host":     "localhost:6379",
	"redis_password": "",
}

// GetConfig returns the application configuration, loading from .env if not already loaded
func GetConfig() *Config {
	_ = GenerateEnvFileIfNotExists(defaults["db_path"])

	if conf != nil {
		return conf
	}

	conf = loadConfig()
	validateConfig(conf)

	return conf
}

// loadConfig loads configuration from .env file with proper parsing and defaults
func loadConfig() *Config {
	conf := &Config{}
	envMap := parseEnvFile(conf_path)

	// Map environment variables to config struct with defaults
	conf.DBPath = getEnvValue(envMap, "db_path", defaults["db_path"])
	conf.RedisHost = getEnvValue(envMap, "redis_host", defaults["redis_host"])
	conf.RedisPassword = getEnvValue(envMap, "redis_password", defaults["redis_password"])

	return conf
}

// parseEnvFile reads and parses .env file with proper error handling
func parseEnvFile(path string) map[string]string {
	envMap := make(map[string]string)

	configContent, err := os.ReadFile(path)
	if err != nil {
		log.Printf("WARNING: Error reading %s file: %v (using defaults)", path, err)
		return envMap
	}

	lines := strings.Split(string(configContent), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by first = only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("WARNING: Invalid config line (missing =): %s", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip quotes from value
		value = strings.Trim(value, "\"'")
		value = strings.TrimSpace(value)

		// Validate key and value
		if key == "" {
			log.Printf("WARNING: Empty key in config")
			continue
		}

		envMap[key] = value
	}

	return envMap
}

// getEnvValue retrieves a value from the env map with a default fallback
func getEnvValue(envMap map[string]string, key, defaultValue string) string {
	if value, exists := envMap[key]; exists && value != "" {
		return value
	}
	return defaultValue
}

// validateConfig validates that required configuration values are set
func validateConfig(conf *Config) {
	if conf.DBPath == "" {
		log.Fatal("ERROR: db_path is required in .env file or defaults")
	}

	log.Printf("INFO: Configuration loaded successfully")
	log.Printf("INFO: Database path: %s", conf.DBPath)
}

func SetConfig(init_conf *Config) {
	conf = init_conf
}

func GenerateEnvFileIfNotExists(dbPath string) error {
	_, err := os.Stat(".env")
	if err == nil {
		return nil
	}
	envContent := fmt.Sprintf("db_path=\"%s\"\n", dbPath)
	return os.WriteFile(".env", []byte(envContent), 0644)
}