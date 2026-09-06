package meta

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEventHandler(t *testing.T) {
	const secret = "app-secret"

	body := []byte(`{"object":"whatsapp_business_account","entry":[]}`)
	handler := NewEventHandler(fakeConfig{metaSecret: secret})

	tests := []struct {
		name       string
		header     func() (key string, value string, set bool)
		body       []byte
		wantStatus int
	}{
		{
			name: "valid signature is accepted and logged",
			header: func() (string, string, bool) {
				return signatureHeader, sign(secret, body), true
			},
			body:       body,
			wantStatus: http.StatusOK,
		},
		{
			name: "tampered signature is unauthorized",
			header: func() (string, string, bool) {
				return signatureHeader, sign("wrong-secret", body), true
			},
			body:       body,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing signature header is bad request",
			header:     func() (string, string, bool) { return "", "", false },
			body:       body,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty body is bad request",
			header: func() (string, string, bool) {
				return signatureHeader, sign(secret, []byte{}), true
			},
			body:       []byte{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
			defer slog.SetDefault(previous)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(tt.body))
			if key, value, set := tt.header(); set {
				request.Header.Set(key, value)
			}

			handler.Handle(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, recorder.Code)
			}

			receivedLogged := bytes.Contains(logs.Bytes(), []byte("meta webhook received"))
			if tt.wantStatus == http.StatusOK && !receivedLogged {
				t.Fatalf("expected payload log for 200, got %q", logs.String())
			}
			if tt.wantStatus != http.StatusOK && receivedLogged {
				t.Fatalf("expected no payload log for %d, got %q", tt.wantStatus, logs.String())
			}

			if tt.wantStatus == http.StatusOK {
				payload := findLogPayload(t, &logs)
				if payload != string(tt.body) {
					t.Fatalf("expected payload %q, got %q", string(tt.body), payload)
				}
			}
		})
	}
}

// findLogPayload scans JSON log lines looking for the debug payload entry.
func findLogPayload(t *testing.T, logs *bytes.Buffer) string {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(logs.Bytes()))
	for decoder.More() {
		var entry map[string]any
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("decode log entry: %v", err)
		}
		if entry["msg"] == "meta webhook payload" {
			payload, _ := entry["payload"].(string)
			return payload
		}
	}
	t.Fatalf("no %q entry found in logs %q", "meta webhook payload", logs.String())
	return ""
}
