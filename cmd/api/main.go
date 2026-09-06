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
	"github.com/jdgonzalez907/1channel/internal/http/middleware"
	"github.com/jdgonzalez907/1channel/internal/logger"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/meta"
	"github.com/jdgonzalez907/1channel/internal/nats"
)

const (
	shutdownTimeout = 10 * time.Second
	requestTimeout  = 10 * time.Second
)

func main() {
	cfg, err := config.NewConfiguration()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger.New(cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	natsClient, err := nats.New(cfg, slog.Default())
	if err != nil {
		slog.Error("failed to connect to nats", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := natsClient.Drain(); err != nil {
			slog.Warn("draining nats connection", "error", err)
		}
	}()

	if err := natsClient.Register(ctx); err != nil {
		slog.Error("failed to register nats streams and consumers", "error", err)
		os.Exit(1)
	}

	router := httprouter.NewRouter()
	router.Use(
		middleware.Recovery(),
		middleware.Logging(),
		middleware.Timeout(requestTimeout),
	)
	meta.RegisterRoutes(router, cfg, natsClient.Streams.MessageEventReceived)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort(),
		Handler: router,
	}

	go func() {
		slog.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	stop()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
