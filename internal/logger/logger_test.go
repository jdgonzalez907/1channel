package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		title           string
		setup           func(t *testing.T, m *config.MockConfiguration)
		expInfoEnabled  bool
		expErrorEnabled bool
	}{
		{
			title: "success - warn level disables info and keeps error",
			setup: func(t *testing.T, m *config.MockConfiguration) {
				t.Helper()
				m.On("LogLevel").Return("warn").Once()
			},
			expInfoEnabled:  false,
			expErrorEnabled: true,
		},
		{
			title: "success - debug level enables info",
			setup: func(t *testing.T, m *config.MockConfiguration) {
				t.Helper()
				m.On("LogLevel").Return("debug").Once()
			},
			expInfoEnabled:  true,
			expErrorEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &config.MockConfiguration{}
			tt.setup(t, m)

			// Act
			logger := New(m)

			// Assert
			ctx := context.Background()
			assert.Equal(t, tt.expInfoEnabled, logger.Enabled(ctx, slog.LevelInfo))
			assert.Equal(t, tt.expErrorEnabled, logger.Enabled(ctx, slog.LevelError))
			m.AssertExpectations(t)
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		title    string
		input    string
		expected slog.Level
	}{
		{title: "success - parses debug", input: "debug", expected: slog.LevelDebug},
		{title: "success - parses uppercase info case-insensitively", input: "INFO", expected: slog.LevelInfo},
		{title: "success - parses warn", input: "warn", expected: slog.LevelWarn},
		{title: "success - parses error", input: "error", expected: slog.LevelError},
		{title: "failure - unknown level falls back to info", input: "verbose", expected: slog.LevelInfo},
		{title: "failure - empty level falls back to info", input: "", expected: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange

			// Act
			got := parseLevel(tt.input)

			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}
