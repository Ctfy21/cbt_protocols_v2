package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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

	// Local API version
	LocalAPIversion int

	// NTP configuration
	NTPLocation     string
	NTPEnabled      bool
	NTPServers      []string
	NTPSyncInterval time.Duration
	NTPTimeout      time.Duration

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
		LocalAPIversion:    getEnvAsInt("LOCAL_API_VERSION", 1),
		ChamberSuffixes:    parseChamberSuffixes(getEnv("CHAMBER_SUFFIXES", "room1,room2,room3,galo,sb4,oreol,sb1")),
		HeartbeatInterval:  getEnvAsInt("HEARTBEAT_INTERVAL", 30),

		// NTP configuration
		NTPEnabled:      getEnvAsBool("NTP_ENABLED", true),
		NTPServers:      parseNTPServers(getEnv("NTP_SERVERS", "ru.pool.ntp.org,europe.pool.ntp.org,0.ru.pool.ntp.org,1.ru.pool.ntp.org,pool.ntp.org")),
		NTPSyncInterval: getEnvAsDuration("NTP_SYNC_INTERVAL", "5m"),
		NTPTimeout:      getEnvAsDuration("NTP_TIMEOUT", "5s"),
		NTPLocation:     getEnv("NTP_LOCATION", "Europe/Moscow"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
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
	log.Printf("NTP configuration: enabled=%v, servers=%v, sync_interval=%v",
		cfg.NTPEnabled, cfg.NTPServers, cfg.NTPSyncInterval)

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

// parseNTPServers parses comma-separated NTP servers
func parseNTPServers(serversStr string) []string {
	if serversStr == "" {
		return []string{"pool.ntp.org"} // Default fallback
	}

	servers := strings.Split(serversStr, ",")
	for i := range servers {
		servers[i] = strings.TrimSpace(servers[i])
	}

	return servers
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

// getEnvAsBool gets an environment variable as boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration gets an environment variable as duration with a default value
func getEnvAsDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}

	// Fallback parsing for defaultValue
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}

	return 5 * time.Minute // Ultimate fallback
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
