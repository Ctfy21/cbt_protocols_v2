package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// MongoDB
	MongoURI      string
	MongoDatabase string

	// Server
	Port    string
	GinMode string

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// API
	APIKey string

	// Chamber
	HeartbeatTimeout time.Duration
	CleanupInterval  time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGODB_DATABASE", "cbt_protocols_v2"),
		Port:          getEnv("PORT", "8080"),
		GinMode:       getEnv("GIN_MODE", "debug"),
		JWTSecret:     getEnv("JWT_SECRET", "default-secret-key"),
		APIKey:        getEnv("API_KEY", ""),
	}

	// Parse JWT expiration
	jwtExp := getEnv("JWT_EXPIRATION", "24h")
	duration, err := time.ParseDuration(jwtExp)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION: %v", err)
	}
	cfg.JWTExpiration = duration

	// Parse heartbeat timeout
	heartbeatTimeout := getEnvInt("HEARTBEAT_TIMEOUT", 300)
	cfg.HeartbeatTimeout = time.Duration(heartbeatTimeout) * time.Second

	// Parse cleanup interval
	cleanupInterval := getEnvInt("CLEANUP_INTERVAL", 300)
	cfg.CleanupInterval = time.Duration(cleanupInterval) * time.Second

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
