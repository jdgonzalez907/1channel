package application

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewReceiveContactMessage(t *testing.T) {
	// Arrange
	repository := &domain.MockConversationRepository{}
	contactsAPI := &contacts.MockContactsAPI{}

	// Act
	useCase := NewReceiveContactMessage(repository, contactsAPI)

	// Assert
	require.NotNil(t, useCase)
}

func TestReceiveContactMessageExecute(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	receivedAt := createdAt.Add(time.Hour)
	conversationID := uuid.NewV7()
	messageID := uuid.NewV7()
	contactID := uuid.NewV7()
	externalContactID := "wa-contact-1"
	externalID := "wa-inbound-1"

	newConversation := func(t *testing.T, status domain.ConversationStatus, unreadCount int8, finishedAt *time.Time, messages map[uuid.UUID]*domain.Message) *domain.Conversation {
		t.Helper()
		conversation, err := domain.NewConversation(conversationID, status, messages, unreadCount, nil, contactID, createdAt, nil, finishedAt)
		require.NoError(t, err)
		return conversation
	}

	newKnownMessage := func(t *testing.T) *domain.Message {
		t.Helper()
		message, err := domain.NewMessage(uuid.NewV7(), &externalID, "hello", domain.Delivered, nil, &contactID, createdAt, nil, nil, nil)
		require.NoError(t, err)
		return message
	}

	input := ReceiveContactMessageInput{
		ConversationID:    conversationID,
		MessageID:         messageID,
		ExternalMessageID: externalID,
		ExternalContactID: externalContactID,
		Text:              "hello",
		ReceivedAt:        receivedAt,
	}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI)
		input         ReceiveContactMessageInput
		expectedError string
	}{
		{
			title: "success - receives message into existing open conversation",
			setup: func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI) {
				t.Helper()
				c.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				conversation := newConversation(t, domain.Pending, 0, nil, nil)
				m.On("FindLastOpenByContactID", mock.Anything, contactID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, mock.MatchedBy(func(saved *domain.Conversation) bool {
					return saved != nil && saved.ID() == conversationID && saved.UnreadCount() == 1
				})).Return(nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "success - opens a new conversation when contact has none open",
			setup: func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI) {
				t.Helper()
				c.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				m.On("FindLastOpenByContactID", mock.Anything, contactID).Return(nil, nil).Once()
				m.On("Save", mock.Anything, mock.MatchedBy(func(saved *domain.Conversation) bool {
					return saved != nil && saved.ID() == conversationID && saved.Status() == domain.Pending && saved.UnreadCount() == 1
				})).Return(nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "success - silently ignores an already received external id",
			setup: func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI) {
				t.Helper()
				c.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				message := newKnownMessage(t)
				conversation := newConversation(t, domain.Pending, 0, nil, map[uuid.UUID]*domain.Message{message.ID(): message})
				m.On("FindLastOpenByContactID", mock.Anything, contactID).Return(conversation, nil).Once()
			},
			input:         input,
			expectedError: "",
		},
		{
			title: "failure - wraps contacts API error",
			setup: func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI) {
				t.Helper()
				c.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(uuid.Nil(), errors.New("contacts service unavailable")).Once()
			},
			input:         input,
			expectedError: "error receiving contact message\ncontacts service unavailable",
		},
		{
			title: "failure - wraps repository find error",
			setup: func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI) {
				t.Helper()
				c.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				m.On("FindLastOpenByContactID", mock.Anything, contactID).Return(nil, errors.New("db connection lost")).Once()
			},
			input:         input,
			expectedError: "error receiving contact message\ndb connection lost",
		},
		{
			title: "failure - wraps closed conversation error",
			setup: func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI) {
				t.Helper()
				c.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				finished := createdAt.Add(-time.Hour)
				conversation := newConversation(t, domain.Expired, 0, &finished, nil)
				m.On("FindLastOpenByContactID", mock.Anything, contactID).Return(conversation, nil).Once()
			},
			input:         input,
			expectedError: "error receiving contact message\n" + domain.ErrConversationClosed.Error(),
		},
		{
			title: "failure - wraps repository save error",
			setup: func(t *testing.T, m *domain.MockConversationRepository, c *contacts.MockContactsAPI) {
				t.Helper()
				c.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				conversation := newConversation(t, domain.Pending, 0, nil, nil)
				m.On("FindLastOpenByContactID", mock.Anything, contactID).Return(conversation, nil).Once()
				m.On("Save", mock.Anything, conversation).Return(errors.New("save failed")).Once()
			},
			input:         input,
			expectedError: "error receiving contact message\nsave failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &domain.MockConversationRepository{}
			c := &contacts.MockContactsAPI{}
			tt.setup(t, m, c)

			// Act
			err := NewReceiveContactMessage(m, c).Execute(context.Background(), tt.input)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			m.AssertExpectations(t)
			c.AssertExpectations(t)
		})
	}
}
