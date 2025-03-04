package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	ServerPort  int
	DatabaseURL string
	Environment  string
	LogLevel    string
}

// Load loads configuration from enviroment variables
func Load() (*Config, error) {
	port := getEnvAsInt("SERVER_PORT", 8080)

	// Database connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvAsInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "postgres")
	dbName := getEnv("DB_NAME", "order_service")
	sslMode := getEnv("DB_SSLMODE", "disable")

	// Construct database URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, sslMode)

	return &Config{
		ServerPort:  port,
		DatabaseURL: dbURL,
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}, nil
}

// Helper functions to read enviroment variables

// getEnv reads an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exits := os.LookupEnv(key); exits {
		return value
	}
	return fallback
}

// getEnvAsInt reads an environment variable as an integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}

	return fallback
}
