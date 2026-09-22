package meta

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfig struct {
	phoneNumberID string
	accessToken   string
}

func (m *mockConfig) AppName() string                        { return "test" }
func (m *mockConfig) HTTPPort() string                       { return "8080" }
func (m *mockConfig) LogLevel() string                       { return "debug" }
func (m *mockConfig) OneChannelSecret() string               { return "secret" }
func (m *mockConfig) MetaSecret() string                     { return "secret" }
func (m *mockConfig) WhatsAppPhoneNumberID() string          { return m.phoneNumberID }
func (m *mockConfig) WhatsAppAccessToken() string            { return m.accessToken }
func (m *mockConfig) PostgresHost() string                   { return "localhost" }
func (m *mockConfig) PostgresPort() uint16                   { return 5432 }
func (m *mockConfig) PostgresDatabase() string               { return "test" }
func (m *mockConfig) PostgresUsername() string               { return "test" }
func (m *mockConfig) PostgresPassword() string               { return "test" }
func (m *mockConfig) PostgresMaxConns() int32                { return 4 }
func (m *mockConfig) PostgresMinConns() int32                { return 0 }
func (m *mockConfig) PostgresMaxConnLifeTime() time.Duration { return time.Hour }
func (m *mockConfig) PostgresMaxConnIdleTime() time.Duration { return 30 * time.Minute }

func TestNewMetaAgentMessageSender(t *testing.T) {
	cfg := &mockConfig{
		phoneNumberID: "123456",
		accessToken:   "test-token",
	}

	sender := NewMetaAgentMessageSender(cfg)
	require.NotNil(t, sender)
}

func TestMetaAgentMessageSenderSend(t *testing.T) {
	messageID := uuid.NewV7()
	to := "16505551234"
	text := "Hello, World!"
	expectedWamid := "wamid.HBgLMTY0NjcwNDM1OTUVAgARGBI1RjQyNUE3NEYxMzAzMzQ5MkEA"

	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	message, err := domain.NewMessage(messageID, nil, text, domain.Registered, nil, nil, createdAt, nil, nil, nil)
	require.NoError(t, err)

	tests := []struct {
		title         string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expectedID    string
		expectedError string
	}{
		{
			title: "success - returns external message id",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var req messageRequest
				err = json.Unmarshal(body, &req)
				require.NoError(t, err)

				assert.Equal(t, "whatsapp", req.MessagingProduct)
				assert.Equal(t, "individual", req.RecipientType)
				assert.Equal(t, to, req.To)
				assert.Equal(t, "text", req.Type)
				assert.Equal(t, text, req.Text.Body)
				assert.Equal(t, messageID.String(), req.BizOpaqueData)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(messageResponse{
					Messages: []struct {
						ID string `json:"id"`
					}{
						{ID: expectedWamid},
					},
				})
			},
			expectedID:    expectedWamid,
			expectedError: "",
		},
		{
			title: "failure - returns error on non-200 status",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "invalid request"}`))
			},
			expectedID:    "",
			expectedError: "unexpected status code: 400, body: {\"error\": \"invalid request\"}",
		},
		{
			title: "failure - returns error on empty messages array",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(messageResponse{
					Messages: []struct {
						ID string `json:"id"`
					}{},
				})
			},
			expectedID:    "",
			expectedError: "no messages in response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			sender := &metaAgentMessageSender{
				client:        server.Client(),
				baseURL:       server.URL,
				phoneNumberID: "123456",
				accessToken:   "test-token",
			}

			wamid, err := sender.Send(context.Background(), to, message)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, wamid)
			}
		})
	}
}
