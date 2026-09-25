package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolveConversation_Execute(t *testing.T) {
	convID := uuid.NewV7()
	contactID := uuid.NewV7()
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()
	now := time.Now()
	finishedAt := now.Add(-time.Hour)
	repoErr := errors.New("repository failure")
	agentsErr := errors.New("agents api failure")

	tests := []struct {
		name       string
		setup      func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation
		wantErr    error
		wantNoSave bool
		check      func(t *testing.T, conv *domain.Conversation)
	}{
		{
			name: "resolves conversation",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.ConversationStatusResolved, conv.Status())
				assert.NotNil(t, conv.FinishedAt())
			},
		},
		{
			name: "agent not owner",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &otherAgentID, &contactID, nil)
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationAgentNotOwner,
		},
		{
			name: "pending without agent",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &contactID, nil)
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationAgentNotOwner,
		},
		{
			name: "already finished is idempotent",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusExpired, &agentID, &contactID, &finishedAt)
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantNoSave: true,
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.ConversationStatusExpired, conv.Status())
			},
		},
		{
			name: "conversation not found",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(nil, nil).Once()
				return nil
			},
			wantErr: domain.ErrConversationNotFound,
		},
		{
			name: "agents api error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(uuid.Nil(), agentsErr).Once()
				return nil
			},
			wantErr: agentsErr,
		},
		{
			name: "find error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(nil, repoErr).Once()
				return nil
			},
			wantErr: repoErr,
		},
		{
			name: "save error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *agents.MockAgentsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(repoErr).Once()
				return conv
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &domain.MockConversationRepository{}
			agentsAPI := &agents.MockAgentsAPI{}
			conv := tt.setup(t, repo, agentsAPI)
			uc := NewResolveConversation(repo, agentsAPI)

			// Act
			err := uc.Execute(context.Background(), ResolveConversationInput{ConversationID: convID, AgentID: agentID, ResolvedAt: now})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrResolvingConversation)
			} else {
				assert.NoError(t, err)
			}
			if tt.check != nil {
				tt.check(t, conv)
			}
			if tt.wantNoSave {
				repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
			}
			repo.AssertExpectations(t)
			agentsAPI.AssertExpectations(t)
		})
	}
}
