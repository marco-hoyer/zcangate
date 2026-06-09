package app

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	// Serial port of the SLCAN adapter (e.g. /dev/ttyACM0).
	SerialPort string
	// Baud rate for the serial connection.
	SerialBaud int
	// TCP address the HTTP server listens on (e.g. :8080).
	HTTPAddr string
	// How long without a measurement before the process is considered dead.
	HeartbeatTimeout time.Duration

	// Bearer token required to execute commands. Empty means auth is disabled.
	CommandAuthToken string

	// InfluxDB connection settings.
	InfluxDBURL      string
	InfluxDBToken    string
	InfluxDBOrg      string
	InfluxDBBucket   string
	InfluxDBLocation string
}

// LoadConfig reads configuration from environment variables, applying defaults
// for any that are not set.
func LoadConfig() Config {
	return Config{
		SerialPort:       getEnv("SERIAL_PORT", "/dev/ttyACM0"),
		SerialBaud:       getEnvInt("SERIAL_BAUD", 115200),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		HeartbeatTimeout: getEnvDuration("HEARTBEAT_TIMEOUT", 60*time.Second),
		CommandAuthToken: os.Getenv("COMMAND_AUTH_TOKEN"),
		InfluxDBURL:      os.Getenv("INFLUXDB_URL"),
		InfluxDBToken:    os.Getenv("INFLUXDB_TOKEN"),
		InfluxDBOrg:      getEnv("INFLUXDB_ORG", "home"),
		InfluxDBBucket:   getEnv("INFLUXDB_BUCKET", "home-metrics"),
		InfluxDBLocation: getEnv("INFLUXDB_LOCATION", "home"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
