package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
)

type fakeConfig struct {
	config.Configuration

	logLevel string
}

func (f fakeConfig) LogLevel() string { return f.logLevel }

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{"debug stays debug", "debug", slog.LevelDebug},
		{"uppercase parses case-insensitively", "INFO", slog.LevelInfo},
		{"warn maps to warn", "warn", slog.LevelWarn},
		{"error maps to error", "error", slog.LevelError},
		{"garbage falls back to info", "verbose", slog.LevelInfo},
		{"empty falls back to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLevel(tt.input); got != tt.want {
				t.Fatalf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewBuildsLoggerHonouringConfiguredLevel(t *testing.T) {
	cfg := fakeConfig{logLevel: "warn"}

	logger := New(cfg)

	ctx := context.Background()
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("expected logger built with LOG_LEVEL=warn to disable info")
	}
	if !logger.Enabled(ctx, slog.LevelError) {
		t.Fatal("expected logger built with LOG_LEVEL=warn to keep error enabled")
	}
}
