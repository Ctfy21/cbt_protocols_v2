package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the service
type Config struct {
	// Server configuration
	Port    string
	GinMode string

	// Home Assistant configuration
	HomeAssistantURL   string
	HomeAssistantToken string

	// MongoDB configuration
	MongoDBURI      string
	MongoDBDatabase string

	// Backend API configuration
	BackendURL    string
	BackendAPIKey string

	// Chamber configuration
	ChamberName     string
	LocalIP         string
	ChamberSuffixes []string // Поддерживаемые суффиксы камер

	// Heartbeat configuration
	HeartbeatInterval int

	// Logging
	LogLevel string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Port:               getEnv("PORT", "8090"),
		GinMode:            getEnv("GIN_MODE", "release"),
		HomeAssistantURL:   getEnv("HA_URL", ""),
		HomeAssistantToken: getEnv("HA_TOKEN", ""),
		MongoDBURI:         getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase:    getEnv("MONGODB_DATABASE", "local_api_v2"),
		BackendURL:         getEnv("BACKEND_URL", "http://localhost:8080/api"),
		BackendAPIKey:      getEnv("BACKEND_API_KEY", ""),
		ChamberName:        getEnv("CHAMBER_NAME", "Climate Chamber"),
		LocalIP:            getEnv("LOCAL_IP", ""),
		ChamberSuffixes:    parseChamberSuffixes(getEnv("CHAMBER_SUFFIXES", "room1,room2,room3,galo,sb4,oreol,sb1")),
		HeartbeatInterval:  getEnvAsInt("HEARTBEAT_INTERVAL", 30),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}

	// Validate required configuration
	if cfg.HomeAssistantURL == "" {
		log.Fatal("HA_URL is required")
	}
	if cfg.HomeAssistantToken == "" {
		log.Fatal("HA_TOKEN is required")
	}

	// Try to get local IP if not set
	if cfg.LocalIP == "" {
		cfg.LocalIP = getLocalIP()
	}

	log.Printf("Configured chamber suffixes: %v", cfg.ChamberSuffixes)

	return cfg, nil
}

// parseChamberSuffixes parses comma-separated chamber suffixes
func parseChamberSuffixes(suffixesStr string) []string {
	if suffixesStr == "" {
		return []string{}
	}

	suffixes := strings.Split(suffixesStr, ",")
	for i := range suffixes {
		suffixes[i] = strings.TrimSpace(suffixes[i])
	}

	return suffixes
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getLocalIP attempts to get the local IP address
func getLocalIP() string {
	// This is a simplified version. In production, you might want to
	// implement a more sophisticated IP detection
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}
