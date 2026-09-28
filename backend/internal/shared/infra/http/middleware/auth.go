package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"uuid"

	usersdomain "github.com/jdgonzalez907/1channel/internal/modules/users/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
)

type contextKey string

const userIDKey contextKey = "userID"

type UserLookup func(ctx context.Context, id uuid.UUID) error

func Auth(findUser UserLookup) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				httperror.Unauthorized(w, r, "missing or malformed bearer token")
				return
			}

			id, err := uuid.Parse(token)
			if err != nil {
				httperror.Unauthorized(w, r, "invalid bearer token")
				return
			}

			if err := findUser(r.Context(), id); err != nil {
				if errors.Is(err, usersdomain.ErrUserNotFound) {
					httperror.Unauthorized(w, r, "unknown user")
					return
				}

				httperror.Generic(w, r, err)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, id)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)

	return id, ok
}

func RequireUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, ok := UserIDFrom(r.Context())
	if !ok {
		httperror.Unauthorized(w, r, "unauthorized")

		return uuid.Nil(), false
	}

	return id, true
}

func bearerToken(header string) (string, bool) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return "", false
	}

	return fields[1], true
}
