package middleware

import (
	"context"
	"net/http"
	"strings"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
)

type contextKey string

const userIDKey contextKey = "userID"

func Auth(next http.Handler) http.Handler {
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

		ctx := context.WithValue(r.Context(), userIDKey, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)

	return id, ok
}

func bearerToken(header string) (string, bool) {
	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return "", false
	}

	return fields[1], true
}
