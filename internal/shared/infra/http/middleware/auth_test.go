package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	validID := uuid.NewV7()

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantID     uuid.UUID
	}{
		{
			name:       "missing header",
			header:     "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed scheme",
			header:     "Token " + validID.String(),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing token",
			header:     "Bearer",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid uuid",
			header:     "Bearer not-a-uuid",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid bearer",
			header:     "Bearer " + validID.String(),
			wantStatus: http.StatusOK,
			wantID:     validID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID uuid.UUID
			var gotOK bool

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotID, gotOK = UserIDFrom(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/v1/users", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			rec := httptest.NewRecorder()
			Auth(next).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				assert.True(t, gotOK)
				assert.Equal(t, tt.wantID, gotID)
			}
		})
	}
}
