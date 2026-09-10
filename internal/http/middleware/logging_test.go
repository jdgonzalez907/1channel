package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeSingleLog(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	require.NoErrorf(t, json.Unmarshal(logs.Bytes(), &entry), "expected single JSON log entry, got %q", logs.String())
	return entry
}

func TestLogging(t *testing.T) {
	tests := []struct {
		title            string
		handler          http.Handler
		request          func() *http.Request
		expStatus        int
		expRequestBytes  int64
		expResponseBytes int
		expClientIP      string
	}{
		{
			title: "success - records explicit status and response body",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "nope", http.StatusUnauthorized)
			}),
			request:          func() *http.Request { return httptest.NewRequest(http.MethodGet, "/meta/webhook", nil) },
			expStatus:        http.StatusUnauthorized,
			expRequestBytes:  0,
			expResponseBytes: len("nope\n"),
			expClientIP:      "192.0.2.1",
		},
		{
			title: "success - defaults to 200 on implicit write",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			}),
			request:          func() *http.Request { return httptest.NewRequest(http.MethodGet, "/", nil) },
			expStatus:        http.StatusOK,
			expRequestBytes:  0,
			expResponseBytes: len("hello"),
			expClientIP:      "192.0.2.1",
		},
		{
			title:   "success - uses declared content length for request size",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			request: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/meta/webhook", nil)
				request.ContentLength = 1234
				return request
			},
			expStatus:        http.StatusOK,
			expRequestBytes:  1234,
			expResponseBytes: 0,
			expClientIP:      "192.0.2.1",
		},
		{
			title: "success - counts consumed body when chunked",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := io.ReadAll(r.Body); err != nil {
					t.Errorf("read body: %v", err)
				}
			}),
			request: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/meta/webhook", strings.NewReader("0123456789"))
				request.ContentLength = -1
				return request
			},
			expStatus:        http.StatusOK,
			expRequestBytes:  10,
			expResponseBytes: 0,
			expClientIP:      "192.0.2.1",
		},
		{
			title:   "success - prefers CF-Connecting-IP over RemoteAddr",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			request: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/meta/webhook", nil)
				request.Header.Set(cfConnectingIPHeader, "31.13.70.54")
				return request
			},
			expStatus:        http.StatusOK,
			expRequestBytes:  0,
			expResponseBytes: 0,
			expClientIP:      "31.13.70.54",
		},
		{
			title:   "success - falls back to raw remote addr when it has no port",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			request: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/", nil)
				request.RemoteAddr = "203.0.113.9"
				return request
			},
			expStatus:        http.StatusOK,
			expRequestBytes:  0,
			expResponseBytes: 0,
			expClientIP:      "203.0.113.9",
		},
		{
			title: "success - keeps first status when handler writes header twice",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.WriteHeader(http.StatusNotFound)
			}),
			request:          func() *http.Request { return httptest.NewRequest(http.MethodGet, "/", nil) },
			expStatus:        http.StatusCreated,
			expRequestBytes:  0,
			expResponseBytes: 0,
			expClientIP:      "192.0.2.1",
		},
		{
			title: "success - sums response bytes across multiple writes",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
				w.Write([]byte("world"))
			}),
			request:          func() *http.Request { return httptest.NewRequest(http.MethodGet, "/", nil) },
			expStatus:        http.StatusOK,
			expRequestBytes:  0,
			expResponseBytes: len("helloworld"),
			expClientIP:      "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			logs := captureSlogLogs(t)
			recorder := httptest.NewRecorder()
			request := tt.request()

			// Act
			Logging()(tt.handler).ServeHTTP(recorder, request)

			// Assert
			entry := decodeSingleLog(t, logs)
			assert.Equal(t, "http request", entry["msg"])
			assert.InDelta(t, tt.expStatus, entry["status"], 0)
			assert.InDelta(t, float64(tt.expRequestBytes), entry["request_bytes"], 0)
			assert.InDelta(t, tt.expResponseBytes, entry["response_bytes"], 0)
			assert.Equal(t, tt.expClientIP, entry["client_ip"])
		})
	}
}
