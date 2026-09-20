package agents

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/application"
	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAgentsAPI(t *testing.T) {
	// Arrange
	useCase := &application.MockFindAgentByID{}

	// Act
	api := NewAgentsAPI(useCase)

	// Assert
	require.NotNil(t, api)
}

func TestAgentsAPIFindAgentByID(t *testing.T) {
	agentID := uuid.NewV7()
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	agent, err := domain.NewAgent(agentID, "jane.doe", createdAt)
	require.NoError(t, err)

	useCaseInput := application.FindAgentByIDInput{ID: agentID}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *application.MockFindAgentByID)
		id            uuid.UUID
		expected      uuid.UUID
		expectedError string
	}{
		{
			title: "success - returns agent id when use case succeeds",
			setup: func(t *testing.T, m *application.MockFindAgentByID) {
				t.Helper()
				m.On("Execute", mock.Anything, useCaseInput).Return(agent, nil).Once()
			},
			id:            agentID,
			expected:      agentID,
			expectedError: "",
		},
		{
			title: "failure - returns nil id and use case error",
			setup: func(t *testing.T, m *application.MockFindAgentByID) {
				t.Helper()
				m.On("Execute", mock.Anything, useCaseInput).Return(nil, errors.New("boom")).Once()
			},
			id:            agentID,
			expected:      uuid.Nil(),
			expectedError: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &application.MockFindAgentByID{}
			tt.setup(t, m)

			// Act
			got, err := NewAgentsAPI(m).FindAgentByID(context.Background(), tt.id)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Equal(t, tt.expected, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
			m.AssertExpectations(t)
		})
	}
}
