package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Mode   string
	Port   string
	DBPath string

	// Cors related
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func (c Config) IsProd() bool {
	return c.Mode == "production"
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Mode:   getEnv("MODE", "production"),
		Port:   getEnv("PORT", "8000"),
		DBPath: getEnv("DB_PATH", "./db/database.sqlite3"),

		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods: getEnvSlice("ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowedHeaders: getEnvSlice("ALLOWED_HEADERS", []string{"Accept", "Content-Type"}),
		ExposedHeaders: getEnvSlice("EXPOSED_HEADERS", []string{"Link"}),

		AllowCredentials: getEnvBool("ALLOW_CREDENTIALS", false),
		MaxAge:           getEnvInt("MAX_AGE", 300),
	}

	return cfg, nil
}

func (c *Config) ServerAddress() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvSlice(key string, defaultVal []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}

	parts := strings.Split(value, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnvBool(key string, defaultVal bool) bool {
	value := os.Getenv(key)
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	value := os.Getenv(key)
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultVal
}
