package nats

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jdgonzalez907/1channel/internal/config"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewFailsWhenServerUnreachable(t *testing.T) {
	cfg := config.NewMockNatsConfiguration("127.0.0.1", 14222, 200*time.Millisecond)

	if _, err := New(cfg, discardLogger()); err == nil {
		t.Fatal("expected error when server unreachable")
	}
}
