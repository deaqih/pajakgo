package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Application
	AppName string
	AppEnv  string
	AppPort string
	AppURL  string

	// Database
	DBHost            string
	DBPort            string
	DBDatabase        string
	DBUsername        string
	DBPassword        string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret        string
	JWTAccessExpire  time.Duration
	JWTRefreshExpire time.Duration

	// Upload
	UploadMaxSize int
	UploadPath    string

	// Processing
	BatchSize         int
	WorkerConcurrency int

	// Asynq
	AsynqRedisAddr     string
	AsynqRedisPassword string
	AsynqRedisDB       int
}

func Load() (*Config, error) {
	// Load .env file if exists
	// Try to load from current dir first, then parent dirs
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env") // For when running from cmd/web or cmd/worker

	cfg := &Config{
		AppName: getEnv("APP_NAME", "Accounting Web"),
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppURL:  getEnv("APP_URL", "http://localhost:8080"),

		DBHost:            getEnv("DB_HOST", "103.150.101.31"),
		DBPort:            getEnv("DB_PORT", "3306"),
		DBDatabase:        getEnv("DB_DATABASE", "pajakgo"),
		DBUsername:        getEnv("DB_USERNAME", "faqih"),
		DBPassword:        getEnv("DB_PASSWORD", "GiantComputer2025!GC"),
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		RedisHost:     getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		JWTSecret:        getEnv("JWT_SECRET", "change-this-secret-key"),
		JWTAccessExpire:  getEnvAsDuration("JWT_ACCESS_EXPIRE", 24*time.Hour),
		JWTRefreshExpire: getEnvAsDuration("JWT_REFRESH_EXPIRE", 168*time.Hour),

		UploadMaxSize: getEnvAsInt("UPLOAD_MAX_SIZE", 104857600), // 100MB
		UploadPath:    getEnv("UPLOAD_PATH", "./storage/uploads"),

		BatchSize:         getEnvAsInt("BATCH_SIZE", 5000),
		WorkerConcurrency: getEnvAsInt("WORKER_CONCURRENCY", 4),

		AsynqRedisAddr:     getEnv("ASYNQ_REDIS_ADDR", "127.0.0.1:6379"),
		AsynqRedisPassword: getEnv("ASYNQ_REDIS_PASSWORD", ""),
		AsynqRedisDB:       getEnvAsInt("ASYNQ_REDIS_DB", 0),
	}

	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		c.DBUsername,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBDatabase,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
