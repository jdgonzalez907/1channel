package conversations

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewConversationsAPI(t *testing.T) {
	// Arrange
	useCase := application.NewMockReceiveContactMessage()

	// Act
	api := NewConversationsAPI(useCase)

	// Assert
	require.NotNil(t, api)
}

func TestConversationsAPIReceiveContactMessage(t *testing.T) {
	externalMessageID := "wamid.abc123"
	externalContactID := "5491112345678"
	text := "hola"
	receivedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *application.MockReceiveContactMessage)
		input         ReceiveContactMessageInput
		expectedError string
	}{
		{
			title: "success - returns nil when use case succeeds",
			setup: func(t *testing.T, m *application.MockReceiveContactMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(input application.ReceiveContactMessageInput) bool {
					return input.ExternalMessageID == externalMessageID &&
						input.ExternalContactID == externalContactID &&
						input.Text == text &&
						input.ReceivedAt.Equal(receivedAt) &&
						input.ConversationID != uuid.Nil() &&
						input.MessageID != uuid.Nil()
				})).Return(nil).Once()
			},
			input: ReceiveContactMessageInput{
				ExternalMessageID: externalMessageID,
				ExternalContactID: externalContactID,
				Text:              text,
				ReceivedAt:        receivedAt,
			},
			expectedError: "",
		},
		{
			title: "failure - returns error when use case fails",
			setup: func(t *testing.T, m *application.MockReceiveContactMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(errors.New("boom")).Once()
			},
			input: ReceiveContactMessageInput{
				ExternalMessageID: externalMessageID,
				ExternalContactID: externalContactID,
				Text:              text,
				ReceivedAt:        receivedAt,
			},
			expectedError: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := application.NewMockReceiveContactMessage()
			tt.setup(t, m)

			// Act
			err := NewConversationsAPI(m).ReceiveContactMessage(context.Background(), tt.input)

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
