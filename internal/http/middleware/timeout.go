package middleware

import (
	"net/http"
	"time"
)

const timeoutResponseMessage = "timeout"

func Timeout(duration time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, duration, timeoutResponseMessage)
	}
}
