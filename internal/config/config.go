// Package config handles configuration loading from environment variables and .env files.
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Database DatabaseConfig
	Kafka    KafkaConfig
	Server   ServerConfig
	Cache    CacheConfig
}

// KafkaConfig holds configuration for Kafka.
type KafkaConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string
	DLQTopic string
}

// ServerConfig holds configuration for the HTTP server.
type ServerConfig struct {
	Host string
	Port string
}

// CacheConfig holds configuration for the in-memory cache.
type CacheConfig struct {
	TTL             time.Duration
	CleanupInterval time.Duration
}

// DatabaseConfig holds configuration for the database connection.
type DatabaseConfig struct {
	Host       string
	Port       string
	User       string
	Password   string
	DBName     string
	SSLMode    string
	MaxRetries int
	RetryDelay time.Duration
}

// LoadConfig reads configuration from .env file or environment variables.
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			log.Println("No .env file found, using environment variables")
		} else {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	return &Config{
		Database: DatabaseConfig{
			Host:       getEnv("DB_HOST", "localhost"),
			Port:       getEnv("DB_PORT", "5432"),
			User:       getEnv("DB_USER", "postgres"),
			Password:   getEnv("DB_PASSWORD", "postgres"),
			DBName:     getEnv("DB_NAME", "wbtech"),
			SSLMode:    getEnv("DB_SSLMODE", "disable"),
			MaxRetries: getIntEnv("DB_MAX_RETRIES", 5),
			RetryDelay: getDurationEnv("DB_RETRY_DELAY", 2*time.Second),
		},
		Kafka: KafkaConfig{
			Brokers:  []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:    getEnv("KAFKA_TOPIC", "orders"),
			GroupID:  getEnv("KAFKA_GROUP_ID", "orders-service"),
			DLQTopic: getEnv("KAFKA_DLQ_TOPIC", "orders-dlq"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8081"),
		},
		Cache: CacheConfig{
			TTL:             getDurationEnv("CACHE_TTL", 5*time.Minute),
			CleanupInterval: getDurationEnv("CACHE_CLEANUP_INTERVAL", 10*time.Minute),
		},
	}, nil
}

// DSN returns the PostgreSQL Data Source Name.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
		log.Printf("Invalid integer for %s, using default: %v", key, defaultValue)
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration for %s, using default: %v", key, defaultValue)
	}
	return defaultValue
}
