package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/webhooks"
)

func main() {
	ctx := context.Background()

	config.NewLogger()

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("loading config",
			"error", err)
		return
	}

	postgresConn, err := config.NewPostgresConnection(ctx, cfg)
	if err != nil {
		slog.Error("connecting to postgres",
			"error", err)
		return
	}
	defer postgresConn.Close()

	natsConn, _, err := config.NewNatsConnection(ctx, cfg)
	if err != nil {
		slog.Error("connecting to nats",
			"error", err)
		return
	}
	defer natsConn.Close()

	mux := config.NewServer(cfg)

	webhooks.NewMetaWebhooks(cfg, mux)

	slog.Info("listening :" + cfg.HTTPPort)
	err = http.ListenAndServe(":"+cfg.HTTPPort, mux)
	if err != nil {
		slog.Error("listening and serving",
			"error", err)
		return
	}
}
