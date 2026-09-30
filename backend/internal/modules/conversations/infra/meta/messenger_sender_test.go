package meta

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

func newAgentTextMessage(t *testing.T, text string) *domain.Message {
	t.Helper()

	agentID := uuid.NewV7()
	message, err := domain.NewMessage(
		uuid.NewV7(),
		domain.MessageStatusSent,
		domain.MessageTypeText,
		&text,
		&agentID,
		nil,
		nil,
		time.Now().UTC(),
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	return message
}

func TestMessengerSender_Send(t *testing.T) {
	t.Run("sends text and returns message id", func(t *testing.T) {
		var gotMethod, gotPath, gotToken string
		var gotBody sendMessageRequest

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotToken = r.URL.Query().Get("access_token")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"recipient_id":"psid-1","message_id":"mid-123"}`))
		}))
		defer server.Close()

		sender := NewMessengerSender(Config{PageID: "page-1", PageAccessToken: "token-1"})
		sender.baseURL = server.URL

		id, err := sender.Send(context.Background(), "psid-1", newAgentTextMessage(t, "hola"))

		require.NoError(t, err)
		assert.Equal(t, "mid-123", id)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v25.0/page-1/messages", gotPath)
		assert.Equal(t, "token-1", gotToken)
		assert.Equal(t, sendMessageRequest{
			Recipient:     sendRecipient{ID: "psid-1"},
			MessagingType: "RESPONSE",
			Message:       sendContent{Text: "hola"},
		}, gotBody)
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		sender := NewMessengerSender(Config{PageID: "page-1", PageAccessToken: "token-1"})
		sender.baseURL = server.URL

		id, err := sender.Send(context.Background(), "psid-1", newAgentTextMessage(t, "hola"))

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("returns error when message id is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"recipient_id":"psid-1"}`))
		}))
		defer server.Close()

		sender := NewMessengerSender(Config{PageID: "page-1", PageAccessToken: "token-1"})
		sender.baseURL = server.URL

		id, err := sender.Send(context.Background(), "psid-1", newAgentTextMessage(t, "hola"))

		assert.Error(t, err)
		assert.Empty(t, id)
	})
}
