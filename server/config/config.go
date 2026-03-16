package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	DBPath        string `json:"db_path"`
	RedisHost     string `json:"redis_host"`
	RedisPassword string `json:"redis_password"`
}

var conf *Config
var conf_path string = ".env"

func GetConfig() *Config {
	_ = GenerateEnvFileIfNotExists("./sjles-pta-vote.db")

	if conf != nil {
		return conf
	}

	conf = &Config{}

	// TODO: Make this into a ini or toml file
	configContent, err := os.ReadFile(conf_path)
	if err != nil {
		log.Printf("Error reading .env file: %v", err)
		os.Exit(1)
	}

	envVariables := strings.Split(string(configContent), "\n")
	envMap := make(map[string]string)

	// TODO: Better error checking for blank variables
	for _, variable := range envVariables {
		if strings.Contains(variable, "=") {
			splitVariable := strings.Split(variable, "=")
			envMap[splitVariable[0]] = splitVariable[1]
		}
	}

	// TODO: Better mapping of key to json values
	// TODO: Better error checking if values are missing
	// TODO: Default values
	for key, value := range envMap {
		// Strip quotes from value if present
		value = strings.Trim(value, "\"")
		value = strings.TrimSpace(value)

		if strings.Contains(key, "db_path") {
			conf.DBPath = value
		} else if strings.Contains(key, "redis_host") {
			conf.RedisHost = value
		} else if strings.Contains(key, "redis_password") {
			conf.RedisPassword = value
		} else {
			log.Printf("Error, Unknown key value pair: %s = %s", key, value)
		}
	}

	return conf
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