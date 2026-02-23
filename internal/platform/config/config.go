package config

import (
	"fmt"
	"os"
)

type Environment string

const (
	Local      Environment = "local"
	Dev        Environment = "dev"
	Production Environment = "production"
)

type Config struct {
	Environment Environment
	Port        string
	Database    DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func Load() *Config {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	return &Config{
		Environment: Environment(env),
		Port:        getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "papaya_user"),
			Password: getEnv("DB_PASSWORD", "papaya_pass"),
			DBName:   getEnv("DB_NAME", "papaya_payout_engine"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
