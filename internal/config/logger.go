package config

import (
	"log/slog"
	"os"
)

func NewLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
