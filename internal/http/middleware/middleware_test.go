package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureSlogLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &logs
}

func tracker(got *[]string, name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*got = append(*got, "before:"+name)
			next.ServeHTTP(w, r)
			*got = append(*got, "after:"+name)
		})
	}
}

func TestChain(t *testing.T) {
	tests := []struct {
		title    string
		build    func(got *[]string) []Middleware
		expected []string
	}{
		{
			title:    "success - no middlewares passes straight to handler",
			build:    func(*[]string) []Middleware { return nil },
			expected: []string{"handler"},
		},
		{
			title:    "success - single middleware wraps handler",
			build:    func(g *[]string) []Middleware { return []Middleware{tracker(g, "a")} },
			expected: []string{"before:a", "handler", "after:a"},
		},
		{
			title:    "success - two middlewares nest in declaration order",
			build:    func(g *[]string) []Middleware { return []Middleware{tracker(g, "a"), tracker(g, "b")} },
			expected: []string{"before:a", "before:b", "handler", "after:b", "after:a"},
		},
		{
			title: "success - three middlewares keep onion order",
			build: func(g *[]string) []Middleware {
				return []Middleware{tracker(g, "a"), tracker(g, "b"), tracker(g, "c")}
			},
			expected: []string{"before:a", "before:b", "before:c", "handler", "after:c", "after:b", "after:a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			var got []string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, "handler")
			})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			// Act
			Chain(handler, tt.build(&got)...).ServeHTTP(recorder, request)

			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}
