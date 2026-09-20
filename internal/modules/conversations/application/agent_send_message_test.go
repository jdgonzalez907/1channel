package application

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents"
	agentsdomain "github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAgentSendMessage(t *testing.T) {
	// Arrange
	repository := &domain.MockConversationRepository{}
	agentsAPI := &agents.MockAgentsAPI{}

	// Act
	useCase := NewAgentSendMessage(repository, agentsAPI)

	// Assert
	require.NotNil(t, useCase)
}

func TestAgentSendMessageExecute(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	sentAt := createdAt.Add(time.Hour)
	conversationID := uuid.NewV7()
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()
	contactID := uuid.NewV7()
	messageID := uuid.NewV7()

	newAssignedConversation := func(t *testing.T) *domain.Conversation {
		t.Helper()
		conversation, err := domain.NewConversation(conversationID, domain.Assigned, nil, 0, &agentID, contactID, createdAt, nil, nil)
		require.NoError(t, err)
		return conversation
	}

	input := AgentSendMessageInput{
		ConversationID: conversationID,
		MessageID:      messageID,
		AgentID:        agentID,
		Text:           "hello",
		SentAt:         sentAt,
	}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *domain.MockConversationRepository, a *agents.MockAgentsAPI)
		input         AgentSendMessageInput
		expectedError string
	}{
		{
			title: "success - stores agent message and persists conversation",
			setup: func(t *testing.T, m *domain.MockConversationRepository, a *agents.MockAgentsAPI) {
				t.Helper()
				conversation := newAssignedConversation(t)
				m.On("FindByID", mock.Anything, conversationID).Return(conversation, nil).Once()
				a.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				m.On("Save", mock.Anything, mock.MatchedBy(func(saved *domain.Conversation) bool {
					return saved != nil && saved.ID() == conversationID && saved.Status() == domain.Assigned
				})).Return(nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "failure - wraps repository find error",
			setup: func(t *testing.T, m *domain.MockConversationRepository, a *agents.MockAgentsAPI) {
				t.Helper()
				a.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				m.On("FindByID", mock.Anything, conversationID).Return(nil, errors.New("db connection lost")).Once()
			},
			input:         input,
			expectedError: "error sending agent message\ndb connection lost",
		},
		{
			title: "failure - aborts when conversation does not exist",
			setup: func(t *testing.T, m *domain.MockConversationRepository, a *agents.MockAgentsAPI) {
				t.Helper()
				a.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				m.On("FindByID", mock.Anything, conversationID).Return(nil, nil).Once()
			},
			input:         input,
			expectedError: "error sending agent message\n" + domain.ErrConversationNotFound.Error(),
		},
		{
			title: "failure - aborts when agent does not exist",
			setup: func(t *testing.T, m *domain.MockConversationRepository, a *agents.MockAgentsAPI) {
				t.Helper()
				a.On("FindAgentByID", mock.Anything, agentID).Return(uuid.Nil(), agentsdomain.ErrAgentNotFound).Once()
			},
			input:         input,
			expectedError: "error sending agent message\n" + agentsdomain.ErrAgentNotFound.Error(),
		},
		{
			title: "failure - rejects message from a different agent",
			setup: func(t *testing.T, m *domain.MockConversationRepository, a *agents.MockAgentsAPI) {
				t.Helper()
				conversation := newAssignedConversation(t)
				m.On("FindByID", mock.Anything, conversationID).Return(conversation, nil).Once()
				a.On("FindAgentByID", mock.Anything, otherAgentID).Return(otherAgentID, nil).Once()
			},
			input: AgentSendMessageInput{
				ConversationID: conversationID,
				MessageID:      messageID,
				AgentID:        otherAgentID,
				Text:           "hello",
				SentAt:         sentAt,
			},
			expectedError: "error sending agent message\n" + domain.ErrUnauthorizedAgent.Error(),
		},
		{
			title: "failure - wraps repository save error",
			setup: func(t *testing.T, m *domain.MockConversationRepository, a *agents.MockAgentsAPI) {
				t.Helper()
				conversation := newAssignedConversation(t)
				m.On("FindByID", mock.Anything, conversationID).Return(conversation, nil).Once()
				a.On("FindAgentByID", mock.Anything, agentID).Return(agentID, nil).Once()
				m.On("Save", mock.Anything, conversation).Return(errors.New("save failed")).Once()
			},
			input:         input,
			expectedError: "error sending agent message\nsave failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &domain.MockConversationRepository{}
			a := &agents.MockAgentsAPI{}
			tt.setup(t, m, a)

			// Act
			err := NewAgentSendMessage(m, a).Execute(context.Background(), tt.input)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			m.AssertExpectations(t)
			a.AssertExpectations(t)
		})
	}
}
