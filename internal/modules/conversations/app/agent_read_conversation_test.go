package app

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
)

func TestAgentReadConversation_Execute(t *testing.T) {
	convID := uuid.NewV7()
	contactID := uuid.NewV7()
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()
	now := time.Now()
	repoErr := errors.New("repository failure")
	agentsErr := errors.New("agents api failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation
		wantErr error
		check   func(t *testing.T, conv *domain.Conversation)
	}{
		{
			name: "marks unread contact messages",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				unread := buildContactMessage(t, uuid.NewV7(), contactID, nil, nil)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, unread)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithContactUnreadMessagesByID", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				messages := conv.Messages()
				assert.Len(t, messages, 1)
				assert.Equal(t, domain.MessageStatusRead, messages[0].Status())
				assert.NotNil(t, messages[0].ReadAt())
			},
		},
		{
			name: "no unread messages is a no-op",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithContactUnreadMessagesByID", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Empty(t, conv.Messages())
			},
		},
		{
			name: "does not mark agent messages",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				agentMsg := buildAgentMessage(t, uuid.NewV7(), agentID)
				contactMsg := buildContactMessage(t, uuid.NewV7(), contactID, nil, nil)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, agentMsg, contactMsg)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithContactUnreadMessagesByID", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				for _, message := range conv.Messages() {
					if message.AgentID() != nil {
						assert.Equal(t, domain.MessageStatusSent, message.Status())
						assert.Nil(t, message.ReadAt())
					} else {
						assert.Equal(t, domain.MessageStatusRead, message.Status())
					}
				}
			},
		},
		{
			name: "agent not assigned",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &otherAgentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithContactUnreadMessagesByID", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationAgentNotOwner,
		},
		{
			name: "conversation not found",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithContactUnreadMessagesByID", mock.Anything, convID).Return(nil, nil).Once()
				return nil
			},
			wantErr: domain.ErrConversationNotFound,
		},
		{
			name: "agents api error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(uuid.Nil(), agentsErr).Once()
				return nil
			},
			wantErr: agentsErr,
		},
		{
			name: "find error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithContactUnreadMessagesByID", mock.Anything, convID).Return(nil, repoErr).Once()
				return nil
			},
			wantErr: repoErr,
		},
		{
			name: "save error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithContactUnreadMessagesByID", mock.Anything, convID).Return(conv, nil).Once()
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
			agentsAPI := &users.MockUsersAPI{}
			conv := tt.setup(t, repo, agentsAPI)
			uc := NewAgentReadConversation(repo, agentsAPI)

			// Act
			err := uc.Execute(context.Background(), AgentReadConversationInput{ConversationID: convID, AgentID: agentID, ReadAt: now})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrAgentReadingConversation)
			} else {
				assert.NoError(t, err)
			}
			if tt.check != nil {
				tt.check(t, conv)
			}
			repo.AssertExpectations(t)
			agentsAPI.AssertExpectations(t)
		})
	}
}
