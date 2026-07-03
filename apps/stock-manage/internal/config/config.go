package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	AppEnv     string
	ServerPort int

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBCharset  string

	StockCenterBaseURL string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		ServerPort:         getEnvInt("SERVER_PORT", 8082),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnvInt("DB_PORT", 3306),
		DBUser:             getEnv("DB_USER", "root"),
		DBPassword:         getEnv("DB_PASSWORD", "root"),
		DBName:             getEnv("DB_NAME", "stock_manage"),
		DBCharset:          getEnv("DB_CHARSET", "utf8mb4"),
		StockCenterBaseURL: getEnv("STOCK_CENTER_BASE_URL", "http://127.0.0.1:8081"),
	}
}

// DSN returns the MySQL data source name for GORM.
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBCharset,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
