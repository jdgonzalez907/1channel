package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rvr := recover()
				if rvr == nil {
					return
				}

				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				logger.Error("panic recovered",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("request_id", chimiddleware.GetReqID(r.Context())),
					slog.Any("panic", rvr),
					slog.String("stack", string(debug.Stack())),
				)

				httperror.Generic(w, r, fmt.Errorf("panic: %v", rvr))
			}()

			next.ServeHTTP(w, r)
		})
	}
}
