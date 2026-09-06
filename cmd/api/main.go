package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httprouter "github.com/jdgonzalez907/1channel/internal/http"
	"github.com/jdgonzalez907/1channel/internal/logger"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/meta"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.NewConfiguration()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger.New(cfg))

	router := httprouter.NewRouter()
	meta.RegisterRoutes(router, cfg)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort(),
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
