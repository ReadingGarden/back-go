package config

import (
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	AppEnv   string
	GinMode  string
	Port     string
	Log      LogConfig
	Database DatabaseConfig
	Swagger  SwaggerConfig
}

type LogConfig struct {
	Format string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type SwaggerConfig struct {
	Enabled  bool
	Host     string
	BasePath string
}

func Load() Config {
	return Config{
		AppEnv:  getEnv("APP_ENV", "local"),
		GinMode: getEnv("GIN_MODE", gin.DebugMode),
		Port:    getEnv("PORT", "8080"),
		Log: LogConfig{
			Format: getEnv("LOG_FORMAT", "text"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnv("DB_PORT", "3306"),
			Name:            getEnv("DB_NAME", "reading_garden"),
			User:            getEnv("DB_USER", "root"),
			Password:        getEnv("DB_PASSWORD", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 3*time.Minute),
		},
		Swagger: SwaggerConfig{
			Enabled:  getEnvBool("SWAGGER_ENABLED", true),
			Host:     getEnv("SWAGGER_HOST", ""),
			BasePath: getEnv("SWAGGER_BASE_PATH", "/api/v1"),
		},
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
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

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
