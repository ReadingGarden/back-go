package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/ReadingGarden/back-go/internal/config"
	"github.com/ReadingGarden/back-go/internal/db"
	"github.com/ReadingGarden/back-go/internal/db/sqlcgen"
)

func TestOpenMySQLWithDocker(t *testing.T) {
	if os.Getenv("INTEGRATION_MYSQL") != "true" {
		t.Skip("set INTEGRATION_MYSQL=true to run docker mysql integration test")
	}

	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("load .env: %v", err)
	}

	cfg := config.Load()

	sqlDB, err := db.Open(cfg.Database)
	if err != nil {
		t.Fatalf("db.Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("PingContext() error = %v", err)
	}

	queries := sqlcgen.New(sqlDB)
	value, err := queries.Ping(ctx)
	if err != nil {
		t.Fatalf("queries.Ping() error = %v", err)
	}
	if value != 1 {
		t.Fatalf("queries.Ping() = %d, want 1", value)
	}
}
