package agents

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/app"
	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAgentsAPI(t *testing.T) {
	// Arrange
	uc := &app.MockFindAgentByID{}

	// Act
	api := NewAgentsAPI(uc)

	// Assert
	assert.NotNil(t, api)
}

func TestAgentsAPI_FindAgentByID(t *testing.T) {
	agentID := uuid.NewV7()
	now := time.Now()
	agent, err := domain.NewAgent(agentID, now)
	assert.NoError(t, err)
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockFindAgentByID)
		id      uuid.UUID
		want    uuid.UUID
		wantErr error
	}{
		{
			name: "success - returns agent id",
			setup: func(t *testing.T, m *app.MockFindAgentByID) {
				t.Helper()
				m.On("Execute", mock.Anything, app.FindAgentByIDInput{ID: agentID}).Return(agent, nil).Once()
			},
			id:   agentID,
			want: agentID,
		},
		{
			name: "failure - propagates use case error",
			setup: func(t *testing.T, m *app.MockFindAgentByID) {
				t.Helper()
				m.On("Execute", mock.Anything, app.FindAgentByIDInput{ID: agentID}).Return(nil, ucErr).Once()
			},
			id:      agentID,
			want:    uuid.Nil(),
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := &app.MockFindAgentByID{}
			tt.setup(t, m)

			// Act
			got, err := NewAgentsAPI(m).FindAgentByID(context.Background(), tt.id)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			m.AssertExpectations(t)
		})
	}
}
