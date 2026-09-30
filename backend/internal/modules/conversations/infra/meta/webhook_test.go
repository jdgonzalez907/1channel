package meta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
)

const (
	testAppSecret   = "app-secret"
	testVerifyToken = "verify-token"
)

func newTestWebhookHandler(receive *app.MockReceiveContactMessage, edit *app.MockReceiveContactMessageEdit) *WebhookHandler {
	return NewWebhookHandler(receive, edit, Config{AppSecret: testAppSecret, VerifyToken: testVerifyToken}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func newTestWebhookRouter(h *WebhookHandler) chi.Router {
	router := chi.NewRouter()
	h.Register(router)

	return router
}

func signBody(body string) string {
	mac := hmac.New(sha256.New, []byte(testAppSecret))
	mac.Write([]byte(body))

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func doWebhookPost(router chi.Router, body, signature string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/meta", strings.NewReader(body))
	if signature != "" {
		req.Header.Set("X-Hub-Signature-256", signature)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestWebhookHandler_Verification(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid verification returns challenge",
			target:     "/webhooks/meta?hub.mode=subscribe&hub.verify_token=verify-token&hub.challenge=12345",
			wantStatus: http.StatusOK,
			wantBody:   "12345",
		},
		{
			name:       "wrong token is forbidden",
			target:     "/webhooks/meta?hub.mode=subscribe&hub.verify_token=nope&hub.challenge=12345",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "wrong mode is forbidden",
			target:     "/webhooks/meta?hub.mode=unsubscribe&hub.verify_token=verify-token&hub.challenge=12345",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestWebhookRouter(newTestWebhookHandler(&app.MockReceiveContactMessage{}, &app.MockReceiveContactMessageEdit{}))

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestWebhookHandler_Signature(t *testing.T) {
	body := `{"object":"page","entry":[{"messaging":[]}]}`

	t.Run("valid signature is processed", func(t *testing.T) {
		router := newTestWebhookRouter(newTestWebhookHandler(&app.MockReceiveContactMessage{}, &app.MockReceiveContactMessageEdit{}))

		rec := doWebhookPost(router, body, signBody(body))

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing signature is rejected", func(t *testing.T) {
		receive := &app.MockReceiveContactMessage{}
		router := newTestWebhookRouter(newTestWebhookHandler(receive, &app.MockReceiveContactMessageEdit{}))

		rec := doWebhookPost(router, body, "")

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		receive.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	})

	t.Run("invalid signature is rejected", func(t *testing.T) {
		receive := &app.MockReceiveContactMessage{}
		router := newTestWebhookRouter(newTestWebhookHandler(receive, &app.MockReceiveContactMessageEdit{}))

		rec := doWebhookPost(router, body, "sha256=deadbeef")

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		receive.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	})
}

func TestWebhookHandler_MessageReceived(t *testing.T) {
	body := `{"object":"page","entry":[{"id":"PAGE-1","time":1458692752478,"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"PAGE-1"},"timestamp":1458692752478,"message":{"mid":"mid.1","text":"hola"}}]}]}`

	receive := &app.MockReceiveContactMessage{}
	receive.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ReceiveContactMessageInput) bool {
		return in.ConversationID != uuid.Nil() &&
			in.MessageID != uuid.Nil() &&
			in.ExternalContactID == "PSID-1" &&
			in.ExternalMessageID != nil && *in.ExternalMessageID == "mid.1" &&
			in.Text == "hola" &&
			in.ReceivedAt.Equal(time.UnixMilli(1458692752478).UTC())
	})).Return(nil).Once()

	router := newTestWebhookRouter(newTestWebhookHandler(receive, &app.MockReceiveContactMessageEdit{}))

	rec := doWebhookPost(router, body, signBody(body))

	assert.Equal(t, http.StatusOK, rec.Code)
	receive.AssertExpectations(t)
}

func TestWebhookHandler_MessageEdit(t *testing.T) {
	body := `{"object":"page","entry":[{"id":"PAGE-1","time":1458668856463,"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"PAGE-1"},"timestamp":1458668856463,"message_edit":{"mid":"mid.1","text":"hola editado","num_edit":1}}]}]}`

	edit := &app.MockReceiveContactMessageEdit{}
	edit.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ReceiveContactMessageEditInput) bool {
		return in.ExternalContactID == "PSID-1" &&
			in.ExternalMessageID == "mid.1" &&
			in.NewText == "hola editado" &&
			in.EditedAt.Equal(time.UnixMilli(1458668856463).UTC())
	})).Return(nil).Once()

	router := newTestWebhookRouter(newTestWebhookHandler(&app.MockReceiveContactMessage{}, edit))

	rec := doWebhookPost(router, body, signBody(body))

	assert.Equal(t, http.StatusOK, rec.Code)
	edit.AssertExpectations(t)
}

func TestWebhookHandler_IgnoresUnsupportedEvents(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "object is not a page",
			body: `{"object":"instagram","entry":[{"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"IG-1"},"timestamp":1,"message":{"mid":"mid.1","text":"hola"}}]}]}`,
		},
		{
			name: "message is an echo",
			body: `{"object":"page","entry":[{"messaging":[{"sender":{"id":"PAGE-1"},"recipient":{"id":"PSID-1"},"timestamp":1,"message":{"mid":"mid.1","text":"hola","is_echo":true}}]}]}`,
		},
		{
			name: "message has only attachments",
			body: `{"object":"page","entry":[{"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"PAGE-1"},"timestamp":1,"message":{"mid":"mid.1","attachments":[{"type":"image","payload":{"url":"http://x"}}]}}]}]}`,
		},
		{
			name: "read receipt",
			body: `{"object":"page","entry":[{"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"PAGE-1"},"timestamp":1,"read":{"watermark":1}}]}]}`,
		},
		{
			name: "delivery receipt",
			body: `{"object":"page","entry":[{"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"PAGE-1"},"timestamp":1,"delivery":{"watermarks":[1],"seq":1}}]}]}`,
		},
		{
			name: "postback",
			body: `{"object":"page","entry":[{"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"PAGE-1"},"timestamp":1,"postback":{"title":"x","payload":"y"}}]}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receive := &app.MockReceiveContactMessage{}
			edit := &app.MockReceiveContactMessageEdit{}
			router := newTestWebhookRouter(newTestWebhookHandler(receive, edit))

			rec := doWebhookPost(router, tt.body, signBody(tt.body))

			assert.Equal(t, http.StatusOK, rec.Code)
			receive.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
			edit.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
		})
	}
}

func TestWebhookHandler_Failure(t *testing.T) {
	body := `{"object":"page","entry":[{"messaging":[{"sender":{"id":"PSID-1"},"recipient":{"id":"PAGE-1"},"timestamp":1,"message":{"mid":"mid.1","text":"hola"}}]}]}`

	receive := &app.MockReceiveContactMessage{}
	receive.On("Execute", mock.Anything, mock.Anything).Return(errors.New("boom")).Once()

	router := newTestWebhookRouter(newTestWebhookHandler(receive, &app.MockReceiveContactMessageEdit{}))

	rec := doWebhookPost(router, body, signBody(body))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	receive.AssertExpectations(t)
}
