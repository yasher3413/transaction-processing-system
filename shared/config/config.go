package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Database
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// Redis
	RedisHost string
	RedisPort int

	// Kafka
	KafkaBrokers         string
	KafkaTransactionsTopic string
	KafkaDLQTopic        string

	// Service
	APIPort              int
	WorkerConsumerGroup  string
	PublisherInterval    time.Duration
	PublisherBatchSize   int

	// Observability
	JaegerEndpoint string
	LogLevel       string
	Env            string

	// API
	APIKey string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		PostgresHost:          getEnv("POSTGRES_HOST", "postgres"),
		PostgresPort:          getEnvAsInt("POSTGRES_PORT", 5432),
		PostgresUser:          getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:      getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:            getEnv("POSTGRES_DB", "transactions"),
		RedisHost:             getEnv("REDIS_HOST", "redis"),
		RedisPort:             getEnvAsInt("REDIS_PORT", 6379),
		KafkaBrokers:          getEnv("KAFKA_BROKERS", "redpanda:9092"),
		KafkaTransactionsTopic: getEnv("KAFKA_TRANSACTIONS_TOPIC", "transactions"),
		KafkaDLQTopic:         getEnv("KAFKA_DLQ_TOPIC", "transactions.dlq"),
		APIPort:               getEnvAsInt("API_PORT", 8080),
		WorkerConsumerGroup:   getEnv("WORKER_CONSUMER_GROUP", "transaction-workers"),
		PublisherInterval:     getEnvAsDuration("PUBLISHER_INTERVAL", 5*time.Second),
		PublisherBatchSize:    getEnvAsInt("PUBLISHER_BATCH_SIZE", 100),
		JaegerEndpoint:        getEnv("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		Env:                   getEnv("ENV", "development"),
		APIKey:                getEnv("API_KEY", ""),
	}

	return cfg, nil
}

// GetPostgresDSN returns the PostgreSQL connection string
func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDB)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}


