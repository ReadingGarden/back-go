package config_test

import (
	"testing"
	"time"

	"github.com/ReadingGarden/back-go/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("GIN_MODE", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_MAX_OPEN_CONNS", "")
	t.Setenv("DB_MAX_IDLE_CONNS", "")
	t.Setenv("DB_CONN_MAX_LIFETIME", "")
	t.Setenv("SWAGGER_ENABLED", "")
	t.Setenv("SWAGGER_HOST", "")
	t.Setenv("SWAGGER_BASE_PATH", "")

	cfg := config.Load()

	if cfg.AppEnv != "local" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "local")
	}
	if cfg.GinMode != "debug" {
		t.Fatalf("GinMode = %q, want %q", cfg.GinMode, "debug")
	}
	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.Log.Format != "text" {
		t.Fatalf("Log.Format = %q, want %q", cfg.Log.Format, "text")
	}
	if cfg.Database.Host != "127.0.0.1" {
		t.Fatalf("Database.Host = %q, want %q", cfg.Database.Host, "127.0.0.1")
	}
	if cfg.Database.Port != "3306" {
		t.Fatalf("Database.Port = %q, want %q", cfg.Database.Port, "3306")
	}
	if cfg.Database.Name != "reading_garden" {
		t.Fatalf("Database.Name = %q, want %q", cfg.Database.Name, "reading_garden")
	}
	if cfg.Database.MaxOpenConns != 10 {
		t.Fatalf("Database.MaxOpenConns = %d, want %d", cfg.Database.MaxOpenConns, 10)
	}
	if cfg.Database.MaxIdleConns != 5 {
		t.Fatalf("Database.MaxIdleConns = %d, want %d", cfg.Database.MaxIdleConns, 5)
	}
	if cfg.Database.ConnMaxLifetime != 3*time.Minute {
		t.Fatalf("Database.ConnMaxLifetime = %s, want %s", cfg.Database.ConnMaxLifetime, 3*time.Minute)
	}
	if !cfg.Swagger.Enabled {
		t.Fatalf("Swagger.Enabled = false, want true")
	}
	if cfg.Swagger.BasePath != "/api/v1" {
		t.Fatalf("Swagger.BasePath = %q, want %q", cfg.Swagger.BasePath, "/api/v1")
	}
}

func TestLoadOverrides(t *testing.T) {
	envs := map[string]string{
		"APP_ENV":              "production",
		"GIN_MODE":             "release",
		"PORT":                 "9090",
		"LOG_FORMAT":           "json",
		"DB_HOST":              "db.internal",
		"DB_PORT":              "4406",
		"DB_NAME":              "garden_prod",
		"DB_USER":              "reader",
		"DB_PASSWORD":          "secret",
		"DB_MAX_OPEN_CONNS":    "20",
		"DB_MAX_IDLE_CONNS":    "7",
		"DB_CONN_MAX_LIFETIME": "5m",
		"SWAGGER_ENABLED":      "false",
		"SWAGGER_HOST":         "api.example.com",
		"SWAGGER_BASE_PATH":    "/custom",
	}

	for key, value := range envs {
		t.Setenv(key, value)
	}

	cfg := config.Load()

	if cfg.AppEnv != "production" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "production")
	}
	if cfg.GinMode != "release" {
		t.Fatalf("GinMode = %q, want %q", cfg.GinMode, "release")
	}
	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.Log.Format != "json" {
		t.Fatalf("Log.Format = %q, want %q", cfg.Log.Format, "json")
	}
	if cfg.Database.Host != "db.internal" {
		t.Fatalf("Database.Host = %q, want %q", cfg.Database.Host, "db.internal")
	}
	if cfg.Database.Port != "4406" {
		t.Fatalf("Database.Port = %q, want %q", cfg.Database.Port, "4406")
	}
	if cfg.Database.Name != "garden_prod" {
		t.Fatalf("Database.Name = %q, want %q", cfg.Database.Name, "garden_prod")
	}
	if cfg.Database.User != "reader" {
		t.Fatalf("Database.User = %q, want %q", cfg.Database.User, "reader")
	}
	if cfg.Database.Password != "secret" {
		t.Fatalf("Database.Password = %q, want %q", cfg.Database.Password, "secret")
	}
	if cfg.Database.MaxOpenConns != 20 {
		t.Fatalf("Database.MaxOpenConns = %d, want %d", cfg.Database.MaxOpenConns, 20)
	}
	if cfg.Database.MaxIdleConns != 7 {
		t.Fatalf("Database.MaxIdleConns = %d, want %d", cfg.Database.MaxIdleConns, 7)
	}
	if cfg.Database.ConnMaxLifetime != 5*time.Minute {
		t.Fatalf("Database.ConnMaxLifetime = %s, want %s", cfg.Database.ConnMaxLifetime, 5*time.Minute)
	}
	if cfg.Swagger.Enabled {
		t.Fatalf("Swagger.Enabled = true, want false")
	}
	if cfg.Swagger.Host != "api.example.com" {
		t.Fatalf("Swagger.Host = %q, want %q", cfg.Swagger.Host, "api.example.com")
	}
	if cfg.Swagger.BasePath != "/custom" {
		t.Fatalf("Swagger.BasePath = %q, want %q", cfg.Swagger.BasePath, "/custom")
	}
}

func TestLoadInvalidValuesFallback(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "bad")
	t.Setenv("DB_MAX_IDLE_CONNS", "bad")
	t.Setenv("DB_CONN_MAX_LIFETIME", "bad")
	t.Setenv("SWAGGER_ENABLED", "bad")

	cfg := config.Load()

	if cfg.Database.MaxOpenConns != 10 {
		t.Fatalf("Database.MaxOpenConns = %d, want %d", cfg.Database.MaxOpenConns, 10)
	}
	if cfg.Database.MaxIdleConns != 5 {
		t.Fatalf("Database.MaxIdleConns = %d, want %d", cfg.Database.MaxIdleConns, 5)
	}
	if cfg.Database.ConnMaxLifetime != 3*time.Minute {
		t.Fatalf("Database.ConnMaxLifetime = %s, want %s", cfg.Database.ConnMaxLifetime, 3*time.Minute)
	}
	if !cfg.Swagger.Enabled {
		t.Fatalf("Swagger.Enabled = false, want true")
	}
}
