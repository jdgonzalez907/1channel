package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	sharedmiddleware "github.com/jdgonzalez907/1channel/internal/shared/infra/http/middleware"
)

func newRouter(logger *slog.Logger, deps dependencies) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID)
	router.Use(sharedmiddleware.Logging(logger))
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.CleanPath)
	router.Use(chimiddleware.StripSlashes)
	router.Use(sharedmiddleware.Timeout(5 * time.Second))

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/v1", func(r chi.Router) {
		deps.usersHandler.Register(r)

		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.Auth)
			deps.conversationsHandler.Register(r)
		})
	})

	return router
}
