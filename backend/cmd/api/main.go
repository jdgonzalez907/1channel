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

	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
)

func main() {
	logger := newLogger()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgdb.New(ctx, postgresDSN())
	if err != nil {
		logger.Error("connecting to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	server := &http.Server{
		Addr:              ":" + httpPort(),
		Handler:           newRouter(logger, newDependencies(db)),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutting down", slog.String("error", err.Error()))
	}
}
