package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/users/app"
	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

func TestUserWriteHandler_Create(t *testing.T) {
	userID := uuid.NewV7()
	now := time.Now()
	user, err := domain.NewUser(userID, now)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		setup      func(t *testing.T, m *app.MockCreateUser)
		wantStatus int
	}{
		{
			name: "creates user and returns location",
			setup: func(t *testing.T, m *app.MockCreateUser) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(user, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid id from use case returns 422",
			setup: func(t *testing.T, m *app.MockCreateUser) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, domain.ErrUserInvalidID).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "use case error returns 500",
			setup: func(t *testing.T, m *app.MockCreateUser) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "request timed out returns 504",
			setup: func(t *testing.T, m *app.MockCreateUser) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded).Once()
			},
			wantStatus: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := &app.MockCreateUser{}
			tt.setup(t, m)

			router := chi.NewRouter()
			NewUserWriteHandler(m).Register(router)

			req := httptest.NewRequest(http.MethodPost, "/users", nil)
			rec := httptest.NewRecorder()

			// Act
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.wantStatus, rec.Code)
			m.AssertExpectations(t)
		})
	}
}
