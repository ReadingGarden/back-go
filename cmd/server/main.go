package main

import (
	"log/slog"

	"github.com/ReadingGarden/back-go/internal/app"
	"github.com/ReadingGarden/back-go/internal/config"
	applog "github.com/ReadingGarden/back-go/internal/logger"
	"github.com/joho/godotenv"
)

// @title           ReadingGarden API
// @version         1.0
// @description     Gin + sqlc migration backend for ReadingGarden.
// @BasePath        /api/v1
func main() {
	_ = godotenv.Load(".env")

	cfg := config.Load()
	logger := applog.New(cfg)
	slog.SetDefault(logger)

	application, err := app.New(cfg)
	if err != nil {
		logger.Error("initialize app", "error", err)
		return
	}

	if err := application.Run(); err != nil {
		logger.Error("run app", "error", err)
	}
}
