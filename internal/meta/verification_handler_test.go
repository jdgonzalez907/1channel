package meta

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureSlogLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &logs
}

func TestVerificationHandlerHandle(t *testing.T) {
	tests := []struct {
		title     string
		setup     func(t *testing.T, m *config.MockConfiguration)
		query     string
		expStatus int
		expBody   string
	}{
		{
			title: "success - valid subscribe request echoes challenge",
			setup: func(t *testing.T, m *config.MockConfiguration) {
				t.Helper()
				m.On("OneChannelSecret").Return("verify-token").Once()
			},
			query:     "?hub.mode=subscribe&hub.verify_token=verify-token&hub.challenge=1158201444",
			expStatus: http.StatusOK,
			expBody:   "1158201444",
		},
		{
			title: "failure - wrong token is forbidden",
			setup: func(t *testing.T, m *config.MockConfiguration) {
				t.Helper()
				m.On("OneChannelSecret").Return("verify-token").Once()
			},
			query:     "?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=123",
			expStatus: http.StatusForbidden,
			expBody:   "Forbidden\n",
		},
		{
			title: "failure - wrong mode is forbidden",
			setup: func(t *testing.T, m *config.MockConfiguration) {
				t.Helper()
				m.On("OneChannelSecret").Return("verify-token").Once()
			},
			query:     "?hub.mode=unsubscribe&hub.verify_token=verify-token&hub.challenge=123",
			expStatus: http.StatusForbidden,
			expBody:   "Forbidden\n",
		},
		{
			title: "failure - empty query is forbidden",
			setup: func(t *testing.T, m *config.MockConfiguration) {
				t.Helper()
				m.On("OneChannelSecret").Return("verify-token").Once()
			},
			query:     "",
			expStatus: http.StatusForbidden,
			expBody:   "Forbidden\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &config.MockConfiguration{}
			tt.setup(t, m)
			handler := NewVerificationHandler(m)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/meta/webhook"+tt.query, nil)

			// Act
			handler.Handle(recorder, request)

			// Assert
			assert.Equal(t, tt.expStatus, recorder.Code)
			assert.Equal(t, tt.expBody, recorder.Body.String())
			if ct := recorder.Header().Get("Content-Type"); tt.expStatus == http.StatusOK {
				assert.Equal(t, "text/plain", ct)
			}
			m.AssertExpectations(t)
		})
	}
}

func TestNewVerificationHandler(t *testing.T) {
	// Arrange
	m := &config.MockConfiguration{}

	// Act
	handler := NewVerificationHandler(m)

	// Assert
	require.NotNil(t, handler)
}
