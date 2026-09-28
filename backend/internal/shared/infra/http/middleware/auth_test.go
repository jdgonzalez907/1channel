package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	usersdomain "github.com/jdgonzalez907/1channel/internal/modules/users/domain"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	validID := uuid.NewV7()

	tests := []struct {
		name       string
		header     string
		lookupErr  error
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
		{
			name:       "unknown user",
			header:     "Bearer " + validID.String(),
			lookupErr:  usersdomain.ErrUserNotFound,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "lookup failure",
			header:     "Bearer " + validID.String(),
			lookupErr:  assert.AnError,
			wantStatus: http.StatusInternalServerError,
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

			lookup := func(context.Context, uuid.UUID) error { return tt.lookupErr }

			req := httptest.NewRequest(http.MethodPost, "/v1/users", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			rec := httptest.NewRecorder()
			Auth(lookup)(next).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusOK {
				assert.True(t, gotOK)
				assert.Equal(t, tt.wantID, gotID)
			}
		})
	}
}
