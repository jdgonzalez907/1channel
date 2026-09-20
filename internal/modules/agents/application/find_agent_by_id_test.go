package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"uuid"
)

func TestNewFindAgentByID(t *testing.T) {
	// Arrange
	repository := &domain.MockAgentRepository{}

	// Act
	useCase := NewFindAgentByID(repository)

	// Assert
	require.NotNil(t, useCase)
}

func TestFindAgentByIDExecute(t *testing.T) {
	agentID := uuid.NewV7()
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	existingAgent, err := domain.NewAgent(agentID, "jane.doe", createdAt)
	require.NoError(t, err)

	input := FindAgentByIDInput{ID: agentID}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *domain.MockAgentRepository)
		input         FindAgentByIDInput
		expected      *domain.Agent
		expectedError string
	}{
		{
			title: "success - returns agent when found",
			setup: func(t *testing.T, m *domain.MockAgentRepository) {
				t.Helper()
				m.On("FindByID", mock.Anything, agentID).Return(existingAgent, nil).Once()
			},
			input:         input,
			expected:      existingAgent,
			expectedError: "",
		},
		{
			title: "failure - wraps repository find error",
			setup: func(t *testing.T, m *domain.MockAgentRepository) {
				t.Helper()
				m.On("FindByID", mock.Anything, agentID).Return(nil, errors.New("db connection lost")).Once()
			},
			input:         input,
			expected:      nil,
			expectedError: "error finding agent by id\ndb connection lost",
		},
		{
			title: "failure - returns not found when agent is missing",
			setup: func(t *testing.T, m *domain.MockAgentRepository) {
				t.Helper()
				m.On("FindByID", mock.Anything, agentID).Return(nil, nil).Once()
			},
			input:         input,
			expected:      nil,
			expectedError: "error finding agent by id\n" + domain.ErrAgentNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &domain.MockAgentRepository{}
			tt.setup(t, m)

			// Act
			agent, err := NewFindAgentByID(m).Execute(context.Background(), tt.input)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, agent)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, agent)
			}
			m.AssertExpectations(t)
		})
	}
}
