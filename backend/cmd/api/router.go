package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	sharedmiddleware "github.com/jdgonzalez907/1channel/internal/shared/infra/http/middleware"
)

func newRouter(logger *slog.Logger, deps dependencies) http.Handler {
	router := chi.NewRouter()

	router.Use(chimiddleware.RequestID)
	router.Use(sharedmiddleware.Logging(logger))
	router.Use(sharedmiddleware.Recover(logger))
	router.Use(chimiddleware.CleanPath)
	router.Use(chimiddleware.StripSlashes)
	router.Use(sharedmiddleware.Timeout(5 * time.Second))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httperror.NotFound(w, r, "route not found")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httperror.MethodNotAllowed(w, r, "method not allowed")
	})

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/v1", func(r chi.Router) {
		deps.userWriteHandler.Register(r)
		deps.contactWebhookHandler.Register(r)

		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.Auth(deps.userLookup))
			deps.conversationReadHandler.Register(r)
			deps.contactReadHandler.Register(r)
			deps.personalInformationReadHandler.Register(r)
			deps.userReadHandler.Register(r)
			deps.conversationWriteHandler.Register(r)
			deps.personalInformationWriteHandler.Register(r)
		})
	})

	return router
}
