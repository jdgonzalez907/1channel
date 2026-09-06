package logger

import (
	"log/slog"
	"os"

	"github.com/jdgonzalez907/1channel/internal/config"
)

func New(cfg config.Configuration) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     parseLevel(cfg.LogLevel()),
		AddSource: false,
	}))
}

func parseLevel(level string) slog.Level {
	var parsed slog.Level
	if err := parsed.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo
	}
	return parsed
}
