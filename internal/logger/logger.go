package logger

import (
	"log/slog"
	"os"

	"github.com/ReadingGarden/back-go/internal/config"
)

func New(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if cfg.GinMode == "debug" || cfg.GinMode == "test" {
		opts.Level = slog.LevelDebug
	}

	if cfg.Log.Format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
