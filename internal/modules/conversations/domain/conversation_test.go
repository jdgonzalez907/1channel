package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestConversation(
	t *testing.T,
	status ConversationStatus,
	agentID *uuid.UUID,
	unreadCount int8,
	finishedAt *time.Time,
) *Conversation {
	t.Helper()
	conversation, err := NewConversation(uuid.NewV7(), status, nil, unreadCount, agentID, uuid.NewV7(), baseTime, nil, finishedAt)
	require.NoError(t, err)
	return conversation
}

func newTestAgentMessage(t *testing.T, agentID uuid.UUID, status MessageStatus) *Message {
	t.Helper()
	message, err := NewMessage(uuid.NewV7(), nil, "agent text", status, &agentID, nil, baseTime, nil, nil, nil)
	require.NoError(t, err)
	return message
}

func newTestContactMessage(t *testing.T, contactID uuid.UUID, status MessageStatus, externalID string) *Message {
	t.Helper()
	message, err := NewMessage(uuid.NewV7(), &externalID, "contact text", status, nil, &contactID, baseTime, nil, nil, nil)
	require.NoError(t, err)
	return message
}

func TestNewConversation(t *testing.T) {
	agentID := uuid.NewV7()
	updatedAt := baseTime.Add(time.Hour)
	finishedAt := baseTime.Add(2 * time.Hour)

	tests := []struct {
		title       string
		status      ConversationStatus
		agentID     *uuid.UUID
		unreadCount int8
		updatedAt   *time.Time
		finishedAt  *time.Time
	}{
		{
			title:       "success - builds conversation with assigned agent",
			status:      Assigned,
			agentID:     &agentID,
			unreadCount: 5,
			updatedAt:   &updatedAt,
			finishedAt:  &finishedAt,
		},
		{
			title:       "success - builds pending conversation without agent boundary",
			status:      Pending,
			agentID:     nil,
			unreadCount: minUnreadCount,
			updatedAt:   nil,
			finishedAt:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			contactID := uuid.NewV7()
			id := uuid.NewV7()
			messages := map[uuid.UUID]*Message{}

			// Act
			conversation, err := NewConversation(id, tt.status, messages, tt.unreadCount, tt.agentID, contactID, baseTime, tt.updatedAt, tt.finishedAt)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, conversation)
			assert.Equal(t, id, conversation.ID())
			assert.Equal(t, tt.status, conversation.Status())
			assert.Empty(t, conversation.messages)
			assert.Empty(t, conversation.dirtyMessages)
			assert.Equal(t, tt.unreadCount, conversation.UnreadCount())
			assert.Equal(t, tt.agentID, conversation.AgentID())
			assert.Equal(t, contactID, conversation.ContactID())
			assert.Equal(t, baseTime, conversation.CreatedAt())
			assert.Equal(t, tt.updatedAt, conversation.UpdatedAt())
			assert.Equal(t, tt.finishedAt, conversation.FinishedAt())
		})
	}
}

func TestNewConversationWithNilFoundMessages(t *testing.T) {
	// Arrange
	// (inputs are fixed for the single case)

	// Act
	conversation, err := NewConversation(uuid.NewV7(), Pending, nil, 0, nil, uuid.NewV7(), baseTime, nil, nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, conversation)
	assert.NotNil(t, conversation.messages)
	assert.Empty(t, conversation.messages)
}

func TestIsClosed(t *testing.T) {
	past := baseTime.Add(-time.Hour)
	future := baseTime.Add(time.Hour)

	tests := []struct {
		title      string
		status     ConversationStatus
		finishedAt *time.Time
		expected   bool
	}{
		{
			title:      "success - expired conversation finished in the past is closed",
			status:     Expired,
			finishedAt: &past,
			expected:   true,
		},
		{
			title:      "success - resolved conversation finished in the past is closed",
			status:     Resolved,
			finishedAt: &past,
			expected:   true,
		},
		{
			title:      "failure - pending conversation without finish time stays open",
			status:     Pending,
			finishedAt: nil,
			expected:   false,
		},
		{
			title:      "failure - assigned conversation without finish time stays open",
			status:     Assigned,
			finishedAt: nil,
			expected:   false,
		},
		{
			title:      "failure - expired conversation without finish time stays open boundary",
			status:     Expired,
			finishedAt: nil,
			expected:   false,
		},
		{
			title:      "failure - expired conversation finishing in the future is not closed yet",
			status:     Expired,
			finishedAt: &future,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			conversation := newTestConversation(t, tt.status, nil, 0, tt.finishedAt)

			// Act
			closed := conversation.IsClosed(baseTime)

			// Assert
			assert.Equal(t, tt.expected, closed)
		})
	}
}

func TestAgentSendMessage(t *testing.T) {
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()
	messageID := uuid.NewV7()
	closedAt := baseTime.Add(-time.Hour)
	sentAt := baseTime.Add(time.Hour)

	tests := []struct {
		title              string
		convStatus         ConversationStatus
		convAgentID        *uuid.UUID
		finishedAt         *time.Time
		senderID           uuid.UUID
		expectedError      string
		expectedConvStatus ConversationStatus
	}{
		{
			title:              "success - first agent message claims the conversation",
			convStatus:         Pending,
			convAgentID:        nil,
			finishedAt:         nil,
			senderID:           agentID,
			expectedError:      "",
			expectedConvStatus: Assigned,
		},
		{
			title:              "success - assigned agent sends message keeping status",
			convStatus:         Assigned,
			convAgentID:        &agentID,
			finishedAt:         nil,
			senderID:           agentID,
			expectedError:      "",
			expectedConvStatus: Assigned,
		},
		{
			title:              "failure - rejects message on closed conversation",
			convStatus:         Expired,
			convAgentID:        nil,
			finishedAt:         &closedAt,
			senderID:           agentID,
			expectedError:      ErrConversationClosed.Error(),
			expectedConvStatus: Expired,
		},
		{
			title:              "failure - rejects different agent than the assigned one",
			convStatus:         Assigned,
			convAgentID:        &agentID,
			finishedAt:         nil,
			senderID:           otherAgentID,
			expectedError:      ErrUnauthorizedAgent.Error(),
			expectedConvStatus: Assigned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			conversation := newTestConversation(t, tt.convStatus, tt.convAgentID, 0, tt.finishedAt)

			// Act
			_, err := conversation.AgentSendMessage(messageID, tt.senderID, "hello", sentAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Equal(t, tt.expectedConvStatus, conversation.Status())
				assert.NotContains(t, conversation.messages, messageID)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConvStatus, conversation.Status())
			message, ok := conversation.messages[messageID]
			require.True(t, ok)
			assert.Same(t, conversation.dirtyMessages[messageID], message)
			assert.Equal(t, Registered, message.Status())
			assert.Nil(t, message.ExternalID())
			require.NotNil(t, message.AgentID())
			assert.Equal(t, tt.senderID, *message.AgentID())
			assert.Nil(t, message.ContactID())
			require.NotNil(t, conversation.UpdatedAt())
			assert.True(t, sentAt.Equal(*conversation.UpdatedAt()))
		})
	}
}

func TestAgentReadConversation(t *testing.T) {
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()
	readAt := baseTime.Add(time.Hour)

	tests := []struct {
		title              string
		convStatus         ConversationStatus
		assignAgent        bool
		unreadCount        int8
		withContactMessage bool
		withAgentMessage   bool
		markerAgentID      uuid.UUID
		expectedError      string
	}{
		{
			title:              "success - marks contact messages as read and resets unread",
			convStatus:         Assigned,
			assignAgent:        true,
			unreadCount:        1,
			withContactMessage: true,
			withAgentMessage:   true,
			markerAgentID:      agentID,
			expectedError:      "",
		},
		{
			title:              "failure - conversation without agent cannot be read",
			convStatus:         Pending,
			assignAgent:        false,
			unreadCount:        0,
			withContactMessage: false,
			withAgentMessage:   false,
			markerAgentID:      agentID,
			expectedError:      ErrConversationNotAssigned.Error(),
		},
		{
			title:              "failure - different agent cannot read conversation",
			convStatus:         Assigned,
			assignAgent:        true,
			unreadCount:        0,
			withContactMessage: false,
			withAgentMessage:   false,
			markerAgentID:      otherAgentID,
			expectedError:      ErrUnauthorizedAgent.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			var conversationAgentID *uuid.UUID
			if tt.assignAgent {
				conversationAgentID = &agentID
			}
			conversation := newTestConversation(t, tt.convStatus, conversationAgentID, tt.unreadCount, nil)
			contactMessage := newTestContactMessage(t, conversation.contactID, Delivered, "wa-contact-1")
			agentMessage := newTestAgentMessage(t, agentID, Registered)
			if tt.withContactMessage {
				conversation.messages[contactMessage.ID()] = contactMessage
			}
			if tt.withAgentMessage {
				conversation.messages[agentMessage.ID()] = agentMessage
			}

			// Act
			err := conversation.AgentReadConversation(tt.markerAgentID, readAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, Read, contactMessage.Status())
			require.NotNil(t, contactMessage.ReadAt())
			assert.True(t, readAt.Equal(*contactMessage.ReadAt()))
			assert.Same(t, conversation.dirtyMessages[contactMessage.ID()], contactMessage)
			assert.NotContains(t, conversation.dirtyMessages, agentMessage.ID())
			assert.Equal(t, Registered, agentMessage.Status())
			assert.Equal(t, minUnreadCount, conversation.UnreadCount())
			require.NotNil(t, conversation.UpdatedAt())
			assert.True(t, readAt.Equal(*conversation.UpdatedAt()))
		})
	}
}

func TestAgentReadConversationWithoutContactMessages(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	conversation := newTestConversation(t, Assigned, &agentID, 0, nil)
	agentMessage := newTestAgentMessage(t, agentID, Registered)
	conversation.messages[agentMessage.ID()] = agentMessage

	// Act
	err := conversation.AgentReadConversation(agentID, baseTime.Add(time.Hour))

	// Assert
	require.NoError(t, err)
	assert.Empty(t, conversation.dirtyMessages)
	assert.Nil(t, conversation.UpdatedAt())
	assert.Equal(t, minUnreadCount, conversation.UnreadCount())
}

func TestAssignAgentMessageExternalID(t *testing.T) {
	agentID := uuid.NewV7()
	channelID := "wa-channel-1"

	tests := []struct {
		title           string
		messageExists   bool
		currentExternal *string
		expectedError   string
		expectTracked   bool
	}{
		{
			title:           "success - assigns channel id to pending agent message",
			messageExists:   true,
			currentExternal: nil,
			expectedError:   "",
			expectTracked:   true,
		},
		{
			title:           "success - skips message that already has a channel id",
			messageExists:   true,
			currentExternal: &channelID,
			expectedError:   "",
			expectTracked:   false,
		},
		{
			title:           "failure - rejects unknown message id",
			messageExists:   false,
			currentExternal: nil,
			expectedError:   ErrMessageNotFound.Error(),
			expectTracked:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			conversation := newTestConversation(t, Assigned, &agentID, 0, nil)
			message := newTestAgentMessage(t, agentID, Registered)
			message.externalID = tt.currentExternal
			if tt.messageExists {
				conversation.messages[message.ID()] = message
			}

			// Act
			err := conversation.AssignAgentMessageExternalID(message.ID(), channelID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Empty(t, conversation.dirtyMessages)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, message.ExternalID())
			assert.Equal(t, channelID, *message.ExternalID())
			if tt.expectTracked {
				assert.Same(t, conversation.dirtyMessages[message.ID()], message)
			} else {
				assert.Empty(t, conversation.dirtyMessages)
			}
		})
	}
}

func TestReceiveContactMessage(t *testing.T) {
	messageID := uuid.NewV7()
	externalID := "wa-inbound-1"
	receivedAt := baseTime.Add(time.Hour)

	tests := []struct {
		title         string
		convStatus    ConversationStatus
		unreadCount   int8
		duplicate     bool
		otherContact  bool
		expectedError string
	}{
		{
			title:         "success - stores delivered message and counts it as unread",
			convStatus:    Pending,
			unreadCount:   0,
			duplicate:     false,
			otherContact:  false,
			expectedError: "",
		},
		{
			title:         "success - unread count stops at maximum boundary",
			convStatus:    Pending,
			unreadCount:   maxUnreadCount,
			duplicate:     false,
			otherContact:  false,
			expectedError: "",
		},
		{
			title:         "failure - rejects message with duplicated external id",
			convStatus:    Pending,
			unreadCount:   0,
			duplicate:     true,
			otherContact:  false,
			expectedError: ErrMessageAlreadyExists.Error(),
		},
		{
			title:         "failure - rejects message from a different contact",
			convStatus:    Pending,
			unreadCount:   0,
			duplicate:     false,
			otherContact:  true,
			expectedError: ErrUnauthorizedContact.Error(),
		},
		{
			title:         "failure - rejects message on closed conversation",
			convStatus:    Expired,
			unreadCount:   0,
			duplicate:     false,
			otherContact:  false,
			expectedError: ErrConversationClosed.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			var finishedAt *time.Time
			if tt.convStatus == Expired {
				finished := baseTime.Add(-time.Hour)
				finishedAt = &finished
			}
			conversation := newTestConversation(t, tt.convStatus, nil, tt.unreadCount, finishedAt)
			senderContactID := conversation.contactID
			if tt.otherContact {
				senderContactID = uuid.NewV7()
			}
			if tt.duplicate {
				duplicate := newTestContactMessage(t, conversation.contactID, Delivered, externalID)
				conversation.messages[duplicate.ID()] = duplicate
			}

			// Act
			_, err := conversation.ReceiveContactMessage(messageID, senderContactID, externalID, "hello contact", receivedAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.NotContains(t, conversation.messages, messageID)
				return
			}
			require.NoError(t, err)
			message, ok := conversation.messages[messageID]
			require.True(t, ok)
			assert.Same(t, conversation.dirtyMessages[messageID], message)
			assert.Equal(t, Delivered, message.Status())
			require.NotNil(t, message.ExternalID())
			assert.Equal(t, externalID, *message.ExternalID())
			require.NotNil(t, message.ContactID())
			assert.Equal(t, conversation.contactID, *message.ContactID())
			assert.Nil(t, message.AgentID())
			if tt.unreadCount >= maxUnreadCount {
				assert.Equal(t, maxUnreadCount, conversation.UnreadCount())
			} else {
				assert.Equal(t, tt.unreadCount+1, conversation.UnreadCount())
			}
			require.NotNil(t, conversation.UpdatedAt())
			assert.True(t, receivedAt.Equal(*conversation.UpdatedAt()))
		})
	}
}

func TestContactUpdateTextMessage(t *testing.T) {
	externalID := "wa-inbound-1"
	editedAt := baseTime.Add(time.Hour)

	tests := []struct {
		title         string
		convStatus    ConversationStatus
		otherContact  bool
		knownExternal bool
		expectedError string
	}{
		{
			title:         "success - updates text of contact message by external id",
			convStatus:    Assigned,
			knownExternal: true,
			expectedError: "",
		},
		{
			title:         "failure - rejects edit on closed conversation",
			convStatus:    Expired,
			knownExternal: true,
			expectedError: ErrConversationClosed.Error(),
		},
		{
			title:         "failure - rejects edit from a different contact",
			convStatus:    Assigned,
			otherContact:  true,
			knownExternal: true,
			expectedError: ErrUnauthorizedContact.Error(),
		},
		{
			title:         "failure - rejects unknown external message id",
			convStatus:    Assigned,
			knownExternal: false,
			expectedError: ErrMessageNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			var agentID = uuid.NewV7()
			var finishedAt *time.Time
			if tt.convStatus == Expired {
				finished := baseTime.Add(-time.Hour)
				finishedAt = &finished
			}
			conversation := newTestConversation(t, tt.convStatus, &agentID, 0, finishedAt)
			message := newTestContactMessage(t, conversation.contactID, Delivered, externalID)
			if tt.knownExternal {
				conversation.messages[message.ID()] = message
			}
			senderContactID := conversation.contactID
			if tt.otherContact {
				senderContactID = uuid.NewV7()
			}

			// Act
			err := conversation.ContactUpdateTextMessage(senderContactID, externalID, "edited text", editedAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Equal(t, "contact text", message.Text())
				assert.Empty(t, conversation.dirtyMessages)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "edited text", message.Text())
			require.NotNil(t, message.UpdatedAt())
			assert.True(t, editedAt.Equal(*message.UpdatedAt()))
			assert.Same(t, conversation.dirtyMessages[message.ID()], message)
			require.NotNil(t, conversation.UpdatedAt())
			assert.True(t, editedAt.Equal(*conversation.UpdatedAt()))
		})
	}
}

func TestContactDeleteMessage(t *testing.T) {
	externalID := "wa-inbound-1"
	firstDelete := baseTime.Add(-time.Hour)
	deletedAt := baseTime.Add(time.Hour)

	tests := []struct {
		title          string
		convStatus     ConversationStatus
		messageStatus  MessageStatus
		messageDeleted bool
		knownExternal  bool
		otherContact   bool
		unreadCount    int8
		expectedUnread int8
		expectTracked  bool
		expectedError  string
	}{
		{
			title:          "success - deletes unread message and decreases counter",
			convStatus:     Assigned,
			messageStatus:  Delivered,
			knownExternal:  true,
			unreadCount:    1,
			expectedUnread: 0,
			expectTracked:  true,
		},
		{
			title:          "success - deleting already read message keeps counter",
			convStatus:     Assigned,
			messageStatus:  Read,
			knownExternal:  true,
			unreadCount:    3,
			expectedUnread: 3,
			expectTracked:  true,
		},
		{
			title:          "success - counter stays at zero deleting unread message boundary",
			convStatus:     Assigned,
			messageStatus:  Delivered,
			knownExternal:  true,
			unreadCount:    minUnreadCount,
			expectedUnread: minUnreadCount,
			expectTracked:  true,
		},
		{
			title:          "success - skips delete of an already deleted message",
			convStatus:     Assigned,
			messageStatus:  Deleted,
			messageDeleted: true,
			knownExternal:  true,
			unreadCount:    0,
			expectedUnread: 0,
			expectTracked:  false,
		},
		{
			title:         "failure - rejects delete on closed conversation",
			convStatus:    Expired,
			messageStatus: Delivered,
			knownExternal: true,
			unreadCount:   1,
			expectedError: ErrConversationClosed.Error(),
		},
		{
			title:         "failure - rejects delete from a different contact",
			convStatus:    Assigned,
			messageStatus: Delivered,
			knownExternal: true,
			otherContact:  true,
			unreadCount:   1,
			expectedError: ErrUnauthorizedContact.Error(),
		},
		{
			title:         "failure - rejects unknown external message id",
			convStatus:    Assigned,
			messageStatus: Delivered,
			knownExternal: false,
			unreadCount:   0,
			expectedError: ErrMessageNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			var agentID = uuid.NewV7()
			var finishedAt *time.Time
			if tt.convStatus == Expired {
				finished := baseTime.Add(-time.Hour)
				finishedAt = &finished
			}
			conversation := newTestConversation(t, tt.convStatus, &agentID, tt.unreadCount, finishedAt)
			message := newTestContactMessage(t, conversation.contactID, tt.messageStatus, externalID)
			if tt.messageDeleted {
				message.deletedAt = &firstDelete
			}
			if tt.knownExternal {
				conversation.messages[message.ID()] = message
			}
			senderContactID := conversation.contactID
			if tt.otherContact {
				senderContactID = uuid.NewV7()
			}

			// Act
			err := conversation.ContactDeleteMessage(senderContactID, externalID, deletedAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.NotEqual(t, Deleted, message.Status())
				assert.Empty(t, conversation.dirtyMessages)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedUnread, conversation.UnreadCount())
			if tt.expectTracked {
				assert.Same(t, conversation.dirtyMessages[message.ID()], message)
				require.NotNil(t, message.DeletedAt())
				assert.True(t, deletedAt.Equal(*message.DeletedAt()))
				require.NotNil(t, conversation.UpdatedAt())
			} else {
				assert.Empty(t, conversation.dirtyMessages)
				require.NotNil(t, message.DeletedAt())
				assert.True(t, firstDelete.Equal(*message.DeletedAt()))
				assert.Nil(t, conversation.UpdatedAt())
			}
		})
	}
}

func TestMessages(t *testing.T) {
	contactID := uuid.NewV7()

	t.Run("returns empty slice when no messages", func(t *testing.T) {
		conv := newTestConversation(t, Pending, nil, 0, nil)
		assert.Empty(t, conv.Messages())
	})

	t.Run("returns messages sorted by created at", func(t *testing.T) {
		conv := newTestConversation(t, Assigned, nil, 0, nil)

		msg1, _ := NewMessage(uuid.NewV7(), nil, "first", Delivered, nil, &contactID, baseTime.Add(-2*time.Hour), nil, nil, nil)
		msg2, _ := NewMessage(uuid.NewV7(), nil, "third", Delivered, nil, &contactID, baseTime, nil, nil, nil)
		msg3, _ := NewMessage(uuid.NewV7(), nil, "second", Delivered, nil, &contactID, baseTime.Add(-1*time.Hour), nil, nil, nil)

		conv.messages[msg1.ID()] = msg1
		conv.messages[msg2.ID()] = msg2
		conv.messages[msg3.ID()] = msg3

		result := conv.Messages()

		require.Len(t, result, 3)
		assert.Equal(t, "first", result[0].Text())
		assert.Equal(t, "second", result[1].Text())
		assert.Equal(t, "third", result[2].Text())
	})
}

func TestDirtyMessages(t *testing.T) {
	contactID := uuid.NewV7()

	t.Run("returns empty slice when no dirty messages", func(t *testing.T) {
		conv := newTestConversation(t, Pending, nil, 0, nil)
		assert.Empty(t, conv.DirtyMessages())
	})

	t.Run("returns only dirty messages sorted by created at", func(t *testing.T) {
		conv := newTestConversation(t, Assigned, nil, 0, nil)

		dirty1, _ := NewMessage(uuid.NewV7(), nil, "dirty-first", Delivered, nil, &contactID, baseTime.Add(-1*time.Hour), nil, nil, nil)
		dirty2, _ := NewMessage(uuid.NewV7(), nil, "dirty-second", Delivered, nil, &contactID, baseTime, nil, nil, nil)
		clean, _ := NewMessage(uuid.NewV7(), nil, "clean", Delivered, nil, &contactID, baseTime.Add(-2*time.Hour), nil, nil, nil)

		conv.messages[dirty1.ID()] = dirty1
		conv.messages[dirty2.ID()] = dirty2
		conv.messages[clean.ID()] = clean

		conv.dirtyMessages[dirty1.ID()] = dirty1
		conv.dirtyMessages[dirty2.ID()] = dirty2

		result := conv.DirtyMessages()

		require.Len(t, result, 2)
		assert.Equal(t, "dirty-first", result[0].Text())
		assert.Equal(t, "dirty-second", result[1].Text())
	})
}
