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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/webhooks"
)

func main() {
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config.NewLogger()

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("loading config",
			"error", err)
		return
	}

	postgresConn, err := config.NewPostgresConnection(appCtx, cfg)
	if err != nil {
		slog.Error("connecting to postgres",
			"error", err)
		return
	}
	defer postgresConn.Close()

	natsConn, _, err := config.NewNatsConnection(appCtx, cfg)
	if err != nil {
		slog.Error("connecting to nats",
			"error", err)
		return
	}
	defer natsConn.Close()

	mux := config.NewServer(cfg)
	webhooks.NewMetaWebhooks(cfg, mux)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	if err := run(appCtx, srv); err != nil {
		slog.Error("listening and serving",
			"error", err)
	}

	stop()
	shutdown(srv, natsConn, postgresConn)
}

func run(appCtx context.Context, srv *http.Server) error {
	serverErr := make(chan error, 1)

	go func() {
		slog.Info("listening " + srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-appCtx.Done():
		slog.Info("shutdown signal received")
		return nil
	case err := <-serverErr:
		return err
	}
}

func shutdown(srv *http.Server, natsConn *nats.Conn, postgresConn *pgxpool.Pool) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("shutting down")

	shutdownHTTP(shutdownCtx, srv)
	drainNATS(shutdownCtx, natsConn)
	postgresConn.Close()
}

func shutdownHTTP(shutdownCtx context.Context, srv *http.Server) {
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutting down http server",
			"error", err)
		if err := srv.Close(); err != nil {
			slog.Error("closing http server",
				"error", err)
		}
	}
}

func drainNATS(shutdownCtx context.Context, natsConn *nats.Conn) {
	drainErr := make(chan error, 1)

	go func() {
		drainErr <- natsConn.Drain()
	}()

	select {
	case err := <-drainErr:
		if err != nil {
			slog.Error("draining nats",
				"error", err)
			natsConn.Close()
		}
	case <-shutdownCtx.Done():
		slog.Error("draining nats",
			"error", shutdownCtx.Err())
		natsConn.Close()
	}
}
