package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	tests := []struct {
		name          string
		handler       http.Handler
		build         func() *http.Request
		wantStatus    int
		wantReqBytes  int64
		wantRespBytes int
		wantIP        string
	}{
		{
			name: "records status and response body from handler",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "nope", http.StatusUnauthorized)
			}),
			build:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/meta/webhook", nil) },
			wantStatus:    http.StatusUnauthorized,
			wantReqBytes:  0,
			wantRespBytes: len("nope\n"),
			wantIP:        "192.0.2.1",
		},
		{
			name: "defaults to 200 on implicit write",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			}),
			build:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/", nil) },
			wantStatus:    http.StatusOK,
			wantReqBytes:  0,
			wantRespBytes: len("hello"),
			wantIP:        "192.0.2.1",
		},
		{
			name:    "uses declared Content-Length when present",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			build: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/meta/webhook", nil)
				request.ContentLength = 1234
				return request
			},
			wantStatus:    http.StatusOK,
			wantReqBytes:  1234,
			wantRespBytes: 0,
			wantIP:        "192.0.2.1",
		},
		{
			name: "counts consumed body when chunked",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := io.ReadAll(r.Body); err != nil {
					t.Errorf("read body: %v", err)
				}
			}),
			build: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/meta/webhook", strings.NewReader("0123456789"))
				request.ContentLength = -1
				return request
			},
			wantStatus:    http.StatusOK,
			wantReqBytes:  10,
			wantRespBytes: 0,
			wantIP:        "192.0.2.1",
		},
		{
			name:    "prefers CF-Connecting-IP over RemoteAddr",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			build: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/meta/webhook", nil)
				request.Header.Set(cfConnectingIPHeader, "31.13.70.54")
				return request
			},
			wantStatus:    http.StatusOK,
			wantReqBytes:  0,
			wantRespBytes: 0,
			wantIP:        "31.13.70.54",
		},
		{
			name:    "falls back to raw RemoteAddr when it has no port",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			build: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/", nil)
				request.RemoteAddr = "203.0.113.9"
				return request
			},
			wantStatus:    http.StatusOK,
			wantReqBytes:  0,
			wantRespBytes: 0,
			wantIP:        "203.0.113.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
			defer slog.SetDefault(previous)

			Logging()(tt.handler).ServeHTTP(httptest.NewRecorder(), tt.build())

			entry := decodeSingleLog(t, &logs)
			if entry["msg"] != "http request" {
				t.Fatalf("expected msg %q, got %v", "http request", entry["msg"])
			}
			if status, ok := entry["status"].(float64); !ok || int(status) != tt.wantStatus {
				t.Fatalf("expected status %d, got %v", tt.wantStatus, entry["status"])
			}
			if size, ok := entry["request_bytes"].(float64); !ok || int64(size) != tt.wantReqBytes {
				t.Fatalf("expected request_bytes %d, got %v", tt.wantReqBytes, entry["request_bytes"])
			}
			if size, ok := entry["response_bytes"].(float64); !ok || int(size) != tt.wantRespBytes {
				t.Fatalf("expected response_bytes %d, got %v", tt.wantRespBytes, entry["response_bytes"])
			}
			if entry["client_ip"] != tt.wantIP {
				t.Fatalf("expected client_ip %q, got %v", tt.wantIP, entry["client_ip"])
			}
		})
	}
}

func decodeSingleLog(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("expected single JSON log entry, got %q", logs.String())
	}
	return entry
}
