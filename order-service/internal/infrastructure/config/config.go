package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the global application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Kafka       KafkaConfig
	Environment string
	LogLevel    string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	Name          string
	SSLMode       string
	URL           string
	MigrationsPath string
}

type OutboxWorkerConfig struct {
	BatchSize       int
	ProcessInterval time.Duration
	MaxRetries      int
	DeadLetterQueue string
	DeadLetterMaxRetries int
}

// KafkaConfig represents the configuration for Kafka messaging
type KafkaConfig struct {
	// Basic configuration
	Brokers           []string
	ClientID          string
	ConnectionTimeout time.Duration
	
	// Component-specific configurations
	Producer ProducerConfig
	Consumer ConsumerConfig
	Topics   TopicConfig
	Security SecurityConfig
}

// ProducerConfig holds configuration specific to Kafka producers
type ProducerConfig struct {
	RetryMax       int
	RetryBackoff   time.Duration
	MessageTimeout time.Duration
	RequiredAcks   string // "none", "leader", or "all"
}

// ConsumerConfig holds configuration specific to Kafka consumers
type ConsumerConfig struct {
	GroupID           string
	AutoOffsetReset   string // "earliest" or "latest"
	HeartbeatInterval time.Duration
	SessionTimeout    time.Duration
	MaxWaitTime       time.Duration
	MinFetchBytes     int
	MaxFetchBytes     int
}

// TopicConfig holds names of Kafka topics used by the application
type TopicConfig struct {
	OrdersCreated   string
	OrdersUpdated   string
	OrdersCancelled string
	OrderPayments   string
}

// SecurityConfig holds Kafka security configuration
type SecurityConfig struct {
	Enabled       bool
	Protocol      string
	SaslMechanism string
	SaslUsername  string
	SaslPassword  string
}

// Load loads the entire application configuration
func Load() (*Config, error) {
	v := viper.New()

	// Setup default values
	setDefaultValues(v)

	// Configure viper to read environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read from config file
	v.SetConfigName("config")                // name of config file (without extension)
	v.SetConfigType("yaml")                  // YAML by default
	v.AddConfigPath(".")                     // look for config in the working directory
	v.AddConfigPath("./config")              // look for config in ./config/ directory
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Build the configuration structure
	config := &Config{
		Environment: v.GetString("environment"),
		LogLevel:    v.GetString("log.level"),
	}

	// Build server configuration
	config.Server = ServerConfig{
		Port: v.GetInt("server.port"),
	}

	// Build database configuration
	dbConfig := DatabaseConfig{
		Host:           v.GetString("db.host"),
		Port:           v.GetInt("db.port"),
		User:           v.GetString("db.user"),
		Password:       v.GetString("db.pass"),
		Name:           v.GetString("db.name"),
		SSLMode:        v.GetString("db.sslmode"),
		MigrationsPath: v.GetString("migrations.path"),
	}
	
	// Construct database URL
	dbConfig.URL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.SSLMode)
	config.Database = dbConfig

	// Build Kafka configuration
	connectionTimeout, _ := time.ParseDuration(v.GetString("kafka.connection_timeout"))
	retryBackoff, _ := time.ParseDuration(v.GetString("kafka.producer.retry_backoff"))
	messageTimeout, _ := time.ParseDuration(v.GetString("kafka.producer.message_timeout"))
	heartbeatInterval, _ := time.ParseDuration(v.GetString("kafka.consumer.heartbeat_interval"))
	sessionTimeout, _ := time.ParseDuration(v.GetString("kafka.consumer.session_timeout"))
	maxWaitTime, _ := time.ParseDuration(v.GetString("kafka.consumer.max_wait_time"))

	config.Kafka = KafkaConfig{
		Brokers:           strings.Split(v.GetString("kafka.brokers"), ","),
		ConnectionTimeout: connectionTimeout,
		
		Producer: ProducerConfig{
			RetryMax:       v.GetInt("kafka.producer.retry_max"),
			RetryBackoff:   retryBackoff,
			MessageTimeout: messageTimeout,
			RequiredAcks:   v.GetString("kafka.producer.required_acks"),
		},
		
		Consumer: ConsumerConfig{
			GroupID:           v.GetString("kafka.consumer.group_id"),
			AutoOffsetReset:   v.GetString("kafka.consumer.auto_offset_reset"),
			HeartbeatInterval: heartbeatInterval,
			SessionTimeout:    sessionTimeout,
			MaxWaitTime:       maxWaitTime,
			MinFetchBytes:     v.GetInt("kafka.consumer.min_fetch_bytes"),
			MaxFetchBytes:     v.GetInt("kafka.consumer.max_fetch_bytes"),
		},
		
		Topics: TopicConfig{
			OrdersCreated:   v.GetString("kafka.topics.orders_created"),
			OrdersUpdated:   v.GetString("kafka.topics.orders_updated"),
			OrdersCancelled: v.GetString("kafka.topics.orders_cancelled"),
			OrderPayments:   v.GetString("kafka.topics.order_payments"),
		},
		
		Security: SecurityConfig{
			Enabled:       v.GetBool("kafka.security.enabled"),
			Protocol:      v.GetString("kafka.security.protocol"),
			SaslMechanism: v.GetString("kafka.security.sasl_mechanism"),
			SaslUsername:  v.GetString("kafka.security.sasl_username"),
			SaslPassword:  v.GetString("kafka.security.sasl_password"),
		},
	}

	return config, nil
}

// setDefaultValues sets all the default configuration values
func setDefaultValues(v *viper.Viper) {
	// Core application defaults
	v.SetDefault("environment", "development")
	v.SetDefault("log.level", "info")
	
	// Server defaults
	v.SetDefault("server.port", 8089)
	
	// Database defaults
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.user", "postgres")
	v.SetDefault("db.pass", "postgres")
	v.SetDefault("db.name", "order_service")
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("migrations.path", "db/migrations")
	
	// Kafka defaults - basic
	v.SetDefault("kafka.brokers", "localhost:9092")
	v.SetDefault("kafka.connection_timeout", "10s")
	
	// Kafka defaults - producer
	v.SetDefault("kafka.producer.retry_max", 3)
	v.SetDefault("kafka.producer.retry_backoff", "100ms")
	v.SetDefault("kafka.producer.message_timeout", "5s")
	v.SetDefault("kafka.producer.required_acks", "all")
	
	// Kafka defaults - consumer
	v.SetDefault("kafka.consumer.group_id", "order-service-group")
	v.SetDefault("kafka.consumer.auto_offset_reset", "earliest")
	v.SetDefault("kafka.consumer.heartbeat_interval", "3s")
	v.SetDefault("kafka.consumer.session_timeout", "30s")
	v.SetDefault("kafka.consumer.max_wait_time", "1s")
	v.SetDefault("kafka.consumer.min_fetch_bytes", 1)
	v.SetDefault("kafka.consumer.max_fetch_bytes", 1048576) // 1MB
	
	// Kafka defaults - topics
	v.SetDefault("kafka.topics.orders_created", "orders-created")
	v.SetDefault("kafka.topics.orders_updated", "orders-updated")
	v.SetDefault("kafka.topics.orders_cancelled", "orders-cancelled")
	v.SetDefault("kafka.topics.order_payments", "order-payments")
	
	// Kafka defaults - security
	v.SetDefault("kafka.security.enabled", false)
	v.SetDefault("kafka.security.protocol", "plaintext")
	v.SetDefault("kafka.security.sasl_mechanism", "plain")
	v.SetDefault("kafka.security.sasl_username", "")
	v.SetDefault("kafka.security.sasl_password", "")
}