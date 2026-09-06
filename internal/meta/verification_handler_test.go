package meta

import (
	"github.com/jdgonzalez907/1channel/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerificationHandler(t *testing.T) {
	cfg := config.NewMockSecretsConfiguration("", "verify-token")
	handler := NewVerificationHandler(cfg)

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid subscribe request echoes challenge",
			query:      "?hub.mode=subscribe&hub.verify_token=verify-token&hub.challenge=1158201444",
			wantStatus: http.StatusOK,
			wantBody:   "1158201444",
		},
		{
			name:       "wrong token is forbidden",
			query:      "?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=123",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "wrong mode is forbidden",
			query:      "?hub.mode=unsubscribe&hub.verify_token=verify-token&hub.challenge=123",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "empty query is forbidden",
			query:      "",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/meta/webhook"+tt.query, nil)

			handler.Handle(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, recorder.Code)
			}
			if tt.wantStatus == http.StatusOK && recorder.Body.String() != tt.wantBody {
				t.Fatalf("expected challenge %q, got %q", tt.wantBody, recorder.Body.String())
			}
		})
	}
}
