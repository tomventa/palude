package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application
type Config struct {
	DatabaseURL string
	OllamaURL   string
	Model       string
	MaxAttempts int
}

// Load creates a new Config instance with values from environment variables
// or defaults if not set
func Load() (*Config, error) {
	config := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "user:password@tcp(localhost:3306)/dbname"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		Model:       getEnv("OLLAMA_MODEL", "llama3.2"),
		MaxAttempts: getEnvInt("MAX_ATTEMPTS", 5),
	}

	return config, nil
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns the value of an environment variable as an integer or a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if n, _ := fmt.Sscanf(value, "%d", &intVal); n == 1 {
			return intVal
		}
	}
	return defaultValue
}
