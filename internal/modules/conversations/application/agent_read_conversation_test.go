package application

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAgentReadConversation(t *testing.T) {
	// Arrange
	repository := &domain.MockConversationRepository{}

	// Act
	useCase := NewAgentReadConversation(repository)

	// Assert
	require.NotNil(t, useCase)
}

func TestAgentReadConversationExecute(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	readAt := createdAt.Add(time.Hour)
	conversationID := uuid.NewV7()
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()
	contactID := uuid.NewV7()
	messageID := uuid.NewV7()
	externalID := "wa-inbound-1"

	newConversation := func(t *testing.T, status domain.ConversationStatus, agent *uuid.UUID, unreadCount int8, foundMessages map[uuid.UUID]*domain.Message) *domain.Conversation {
		t.Helper()
		conversation, err := domain.NewConversation(conversationID, status, foundMessages, unreadCount, agent, contactID, createdAt, nil, nil)
		require.NoError(t, err)
		return conversation
	}

	newContactMessage := func(t *testing.T) *domain.Message {
		t.Helper()
		message, err := domain.NewMessage(messageID, &externalID, "hello", domain.Delivered, nil, &contactID, createdAt, nil, nil, nil)
		require.NoError(t, err)
		return message
	}

	input := AgentReadConversationInput{
		ConversationID: conversationID,
		AgentID:        agentID,
		ReadAt:         readAt,
	}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *domain.MockConversationRepository)
		input         AgentReadConversationInput
		expectedError string
	}{
		{
			title: "success - marks unread messages as read and persists conversation",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				message := newContactMessage(t)
				conversation := newConversation(t, domain.Assigned, &agentID, 1, map[uuid.UUID]*domain.Message{messageID: message})
				m.On("FindByID", mock.Anything, conversationID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, mock.MatchedBy(func(saved *domain.Conversation) bool {
					return saved != nil && saved.ID() == conversationID && saved.UnreadCount() == 0
				})).Return(nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "success - silently ignores read of an unassigned conversation",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation := newConversation(t, domain.Pending, nil, 0, nil)
				m.On("FindByID", mock.Anything, conversationID).Return(conversation, nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "failure - wraps repository find error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				m.On("FindByID", mock.Anything, conversationID).Return(nil, errors.New("db connection lost")).Once()
			},
			input:         input,
			expectedError: "error agent reading conversation\ndb connection lost",
		},
		{
			title: "failure - aborts when conversation does not exist",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				m.On("FindByID", mock.Anything, conversationID).Return(nil, nil).Once()
			},
			input:         input,
			expectedError: "error agent reading conversation\n" + domain.ErrConversationNotFound.Error(),
		},
		{
			title: "failure - wraps unauthorized agent error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation := newConversation(t, domain.Assigned, &agentID, 0, nil)
				m.On("FindByID", mock.Anything, conversationID).Return(conversation, nil).Once()
			},
			input: AgentReadConversationInput{
				ConversationID: conversationID,
				AgentID:        otherAgentID,
				ReadAt:         readAt,
			},
			expectedError: "error agent reading conversation\n" + domain.ErrUnauthorizedAgent.Error(),
		},
		{
			title: "failure - wraps repository save error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				message := newContactMessage(t)
				conversation := newConversation(t, domain.Assigned, &agentID, 1, map[uuid.UUID]*domain.Message{messageID: message})
				m.On("FindByID", mock.Anything, conversationID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, conversation).Return(errors.New("save failed")).Once()
			},
			input:         input,
			expectedError: "error agent reading conversation\nsave failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &domain.MockConversationRepository{}
			tt.setup(t, m)

			// Act
			err := NewAgentReadConversation(m).Execute(context.Background(), tt.input)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			m.AssertExpectations(t)
		})
	}
}
