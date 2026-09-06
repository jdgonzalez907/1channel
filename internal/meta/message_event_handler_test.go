package meta

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/nats"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/mock"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return signaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

func newPublishedStream(t *testing.T, publishErr error) (*nats.MessageEventReceivedStream, *nats.MockJetStream) {
	t.Helper()

	js := &nats.MockJetStream{}
	js.On("CreateOrUpdateStream", mock.Anything).Return(&nats.MockStream{}, nil).Once()

	stream, err := nats.NewMessageMessageEventReceivedStream(t.Context(), js)
	if err != nil {
		t.Fatalf("new stream: %v", err)
	}
	js.On("Publish", nats.MessageEventReceivedSubject, mock.Anything).
		Return(&jetstream.PubAck{}, publishErr)
	return stream, js
}

func TestMessageEventHandlerPublishesSignedRequests(t *testing.T) {
	const secret = "app-secret"

	body := []byte(`{"object":"whatsapp_business_account","entry":[]}`)

	tests := []struct {
		name       string
		publishErr error
		wantStatus int
	}{
		{
			name:       "valid signature is published and accepted",
			publishErr: nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "publish failure is internal server error",
			publishErr: errors.New("nats down"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
			defer slog.SetDefault(previous)

			stream, js := newPublishedStream(t, tt.publishErr)
			handler := NewMessageEventHandler(config.NewMockSecretsConfiguration(secret, ""), stream)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/meta/webhook", bytes.NewReader(body))
			request.Header.Set(signatureHeader, sign(secret, body))

			handler.Handle(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, recorder.Code)
			}

			if tt.wantStatus == http.StatusOK {
				js.AssertCalled(t, "Publish", nats.MessageEventReceivedSubject, body)
			} else {
				if !bytes.Contains(logs.Bytes(), []byte("publishing message event to nats")) {
					t.Fatalf("expected publish error log, got %q", logs.String())
				}
			}
		})
	}
}

func TestMessageEventHandlerRejectsRequestsBeforePublishing(t *testing.T) {
	const secret = "app-secret"

	body := []byte(`{"object":"whatsapp_business_account","entry":[]}`)
	handler := NewMessageEventHandler(config.NewMockSecretsConfiguration(secret, ""), nil)

	tests := []struct {
		name       string
		header     func() (key string, value string, set bool)
		body       []byte
		wantStatus int
	}{
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
			if bytes.Contains(logs.Bytes(), []byte("message event received")) {
				t.Fatalf("expected no event log for %d, got %q", tt.wantStatus, logs.String())
			}
		})
	}
}
