package main

import (
	"log/slog"
	"net/http"
	"os"

	httprouter "github.com/jdgonzalez907/1channel/internal/http"
	"github.com/jdgonzalez907/1channel/internal/logger"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/meta"
)

func main() {
	cfg, err := config.NewConfiguration()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger.New(cfg))

	router := httprouter.NewRouter()
	meta.RegisterRoutes(router, cfg)

	addr := ":" + cfg.HTTPPort()
	slog.Info("server starting", "addr", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
