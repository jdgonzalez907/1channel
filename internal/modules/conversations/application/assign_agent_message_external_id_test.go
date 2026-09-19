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

func TestNewAssignAgentMessageExternalID(t *testing.T) {
	// Arrange
	repository := &domain.MockConversationRepository{}

	// Act
	useCase := NewAssignAgentMessageExternalID(repository)

	// Assert
	require.NotNil(t, useCase)
}

func TestAssignAgentMessageExternalIDExecute(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	conversationID := uuid.NewV7()
	messageID := uuid.NewV7()
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	externalID := "wa-channel-1"

	newConversationWithAgentMessage := func(t *testing.T) *domain.Conversation {
		t.Helper()
		message, err := domain.NewMessage(messageID, nil, "hello", domain.Registered, &agentID, nil, createdAt, nil, nil, nil)
		require.NoError(t, err)
		conversation, err := domain.NewConversation(conversationID, domain.Assigned, map[uuid.UUID]*domain.Message{messageID: message}, 0, &agentID, contactID, createdAt, nil, nil)
		require.NoError(t, err)
		return conversation
	}

	input := AssignAgentMessageExternalIDInput{
		MessageID:         messageID,
		ExternalMessageID: externalID,
	}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *domain.MockConversationRepository)
		input         AssignAgentMessageExternalIDInput
		expectedError string
	}{
		{
			title: "success - assigns channel id and persists conversation",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation := newConversationWithAgentMessage(t)
				m.On("FindWithSpecificMessageByMessageID", mock.Anything, messageID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, mock.MatchedBy(func(saved *domain.Conversation) bool {
					return saved != nil && saved.ID() == conversationID
				})).Return(nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "failure - wraps repository find error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				m.On("FindWithSpecificMessageByMessageID", mock.Anything, messageID).Return(nil, errors.New("db connection lost")).Once()
			},
			input:         input,
			expectedError: "error assigning agent message external id\ndb connection lost",
		},
		{
			title: "failure - aborts when conversation does not exist",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				m.On("FindWithSpecificMessageByMessageID", mock.Anything, messageID).Return(nil, nil).Once()
			},
			input:         input,
			expectedError: "error assigning agent message external id\n" + domain.ErrConversationNotFound.Error(),
		},
		{
			title: "failure - wraps unknown message error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation, err := domain.NewConversation(conversationID, domain.Assigned, nil, 0, &agentID, contactID, createdAt, nil, nil)
				require.NoError(t, err)
				m.On("FindWithSpecificMessageByMessageID", mock.Anything, messageID).Return(conversation, nil).Once()
			},
			input:         input,
			expectedError: "error assigning agent message external id\n" + domain.ErrMessageNotFound.Error(),
		},
		{
			title: "failure - wraps repository save error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation := newConversationWithAgentMessage(t)
				m.On("FindWithSpecificMessageByMessageID", mock.Anything, messageID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, conversation).Return(errors.New("save failed")).Once()
			},
			input:         input,
			expectedError: "error assigning agent message external id\nsave failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &domain.MockConversationRepository{}
			tt.setup(t, m)

			// Act
			err := NewAssignAgentMessageExternalID(m).Execute(context.Background(), tt.input)

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
