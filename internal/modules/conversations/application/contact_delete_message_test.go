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

func TestNewContactDeleteMessage(t *testing.T) {
	// Arrange
	repository := &domain.MockConversationRepository{}

	// Act
	useCase := NewContactDeleteMessage(repository)

	// Assert
	require.NotNil(t, useCase)
}

func TestContactDeleteMessageExecute(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	deletedAt := createdAt.Add(time.Hour)
	conversationID := uuid.NewV7()
	messageID := uuid.NewV7()
	contactID := uuid.NewV7()
	externalID := "wa-inbound-1"

	newConversationWithContactMessage := func(t *testing.T) *domain.Conversation {
		t.Helper()
		message, err := domain.NewMessage(messageID, &externalID, "hello", domain.Delivered, nil, &contactID, createdAt, nil, nil, nil)
		require.NoError(t, err)
		conversation, err := domain.NewConversation(conversationID, domain.Assigned, map[uuid.UUID]*domain.Message{messageID: message}, 1, nil, contactID, createdAt, nil, nil)
		require.NoError(t, err)
		return conversation
	}

	input := ContactDeleteMessageInput{
		ExternalMessageID: externalID,
		ContactID:         contactID,
		DeletedAt:         deletedAt,
	}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *domain.MockConversationRepository)
		input         ContactDeleteMessageInput
		expectedError string
	}{
		{
			title: "success - deletes contact message and persists conversation",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation := newConversationWithContactMessage(t)
				m.On("FindWithSpecificMessageByExternalID", mock.Anything, externalID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, mock.MatchedBy(func(saved *domain.Conversation) bool {
					return saved != nil && saved.ID() == conversationID && saved.UnreadCount() == 0
				})).Return(nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "failure - wraps repository find error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				m.On("FindWithSpecificMessageByExternalID", mock.Anything, externalID).Return(nil, errors.New("db connection lost")).Once()
			},
			input:         input,
			expectedError: "error contact deleting text message\ndb connection lost",
		},
		{
			title: "failure - aborts when conversation does not exist",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				m.On("FindWithSpecificMessageByExternalID", mock.Anything, externalID).Return(nil, nil).Once()
			},
			input:         input,
			expectedError: "error contact deleting text message\n" + domain.ErrConversationNotFound.Error(),
		},
		{
			title: "failure - wraps unknown message error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation, err := domain.NewConversation(conversationID, domain.Assigned, nil, 0, nil, contactID, createdAt, nil, nil)
				require.NoError(t, err)
				m.On("FindWithSpecificMessageByExternalID", mock.Anything, externalID).Return(conversation, nil).Once()
			},
			input:         input,
			expectedError: "error contact deleting text message\n" + domain.ErrMessageNotFound.Error(),
		},
		{
			title: "failure - wraps repository save error",
			setup: func(t *testing.T, m *domain.MockConversationRepository) {
				t.Helper()
				conversation := newConversationWithContactMessage(t)
				m.On("FindWithSpecificMessageByExternalID", mock.Anything, externalID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, conversation).Return(errors.New("save failed")).Once()
			},
			input:         input,
			expectedError: "error contact deleting text message\nsave failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &domain.MockConversationRepository{}
			tt.setup(t, m)

			// Act
			err := NewContactDeleteMessage(m).Execute(context.Background(), tt.input)

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
