package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindAgentByID_Execute(t *testing.T) {
	agentID := uuid.NewV7()
	now := time.Now()
	agent, err := domain.NewAgent(agentID, now)
	assert.NoError(t, err)
	repoErr := errors.New("repository failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *domain.MockAgentRepository)
		input   FindAgentByIDInput
		want    *domain.Agent
		wantErr error
	}{
		{
			name: "agent found",
			setup: func(t *testing.T, repo *domain.MockAgentRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, agentID).Return(agent, nil).Once()
			},
			input: FindAgentByIDInput{ID: agentID},
			want:  agent,
		},
		{
			name: "agent not found",
			setup: func(t *testing.T, repo *domain.MockAgentRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, agentID).Return(nil, nil).Once()
			},
			input:   FindAgentByIDInput{ID: agentID},
			wantErr: domain.ErrAgentNotFound,
		},
		{
			name: "repository error",
			setup: func(t *testing.T, repo *domain.MockAgentRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, agentID).Return(nil, repoErr).Once()
			},
			input:   FindAgentByIDInput{ID: agentID},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &domain.MockAgentRepository{}
			tt.setup(t, repo)
			uc := NewFindAgentByID(repo)

			// Act
			got, err := uc.Execute(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrFindingAgentByID)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			repo.AssertExpectations(t)
		})
	}
}
