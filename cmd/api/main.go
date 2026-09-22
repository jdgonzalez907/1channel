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
	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	contactsapp "github.com/jdgonzalez907/1channel/internal/modules/contacts/application"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations"
	convapp "github.com/jdgonzalez907/1channel/internal/modules/conversations/application"
	"github.com/jdgonzalez907/1channel/internal/postgres"
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

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	contactRepo := postgres.NewContactRepository(pool)
	convRepo := postgres.NewConversationRepository(pool)

	getOrCreateContact := contactsapp.NewGetOrCreateContactByExternalID(contactRepo)
	findContactByID := contactsapp.NewFindContactByID(contactRepo)
	contactsAPI := contacts.NewContactsAPI(getOrCreateContact, findContactByID)

	receiveContactMessage := convapp.NewReceiveContactMessage(convRepo, contactsAPI)
	conversationsAPI := conversations.NewConversationsAPI(receiveContactMessage)

	router := httprouter.NewRouter()
	router.Use(
		middleware.Recovery(),
		middleware.Logging(),
		middleware.Timeout(requestTimeout),
	)
	meta.RegisterRoutes(router, cfg, conversationsAPI)

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
