package conversations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestConversation(t *testing.T, status ConversationStatus, found map[string]*Message, unreadCount int) *Conversation {
	t.Helper()
	if found == nil {
		found = map[string]*Message{}
	}
	conversation, err := NewConversation(
		"conv-1",
		status,
		found,
		map[string]*Message{},
		map[string]*Message{},
		map[string]*Message{},
		unreadCount,
		nil,
		"contact-1",
		baseTime,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	return conversation
}

func newTestMessage(t *testing.T, id string, status MessageStatus, updatedAt, deletedAt, readAt *time.Time) *Message {
	t.Helper()
	msg, err := NewMessage(id, "original text", status, false, baseTime, updatedAt, deletedAt, readAt)
	require.NoError(t, err)
	return msg
}

func TestNewConversationStatus(t *testing.T) {
	tests := []struct {
		title    string
		value    string
		expected ConversationStatus
	}{
		{
			title:    "success - converts known status value",
			value:    "pending",
			expected: Pending,
		},
		{
			title:    "success - converts empty value boundary",
			value:    "",
			expected: ConversationStatus(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			status, err := NewConversationStatus(tt.value)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestConversationValue(t *testing.T) {
	tests := []struct {
		title    string
		status   ConversationStatus
		expected string
	}{
		{
			title:    "success - returns string of resolved status",
			status:   Resolved,
			expected: "resolved",
		},
		{
			title:    "success - returns empty string for empty status boundary",
			status:   ConversationStatus(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			value := tt.status.Value()

			// Assert
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestNewConversation(t *testing.T) {
	agentID := "agent-1"
	updatedAt := baseTime.Add(time.Hour)
	finishedAt := baseTime.Add(2 * time.Hour)
	deletedAt := baseTime.Add(3 * time.Hour)
	found := map[string]*Message{"m-1": newTestMessage(t, "m-1", Sent, nil, nil, nil)}
	added := map[string]*Message{"m-2": newTestMessage(t, "m-2", Sent, nil, nil, nil)}
	updated := map[string]*Message{}
	deleted := map[string]*Message{}

	tests := []struct {
		title    string
		agentID  *string
		expected *string
	}{
		{
			title:    "success - builds conversation with assigned agent",
			agentID:  &agentID,
			expected: &agentID,
		},
		{
			title:    "success - builds conversation without agent boundary",
			agentID:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			conversation, err := NewConversation(
				"conv-1",
				Assigned,
				found,
				added,
				updated,
				deleted,
				5,
				tt.agentID,
				"contact-1",
				baseTime,
				&updatedAt,
				&finishedAt,
				&deletedAt,
			)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, conversation)
			assert.Equal(t, "conv-1", conversation.ID())
			assert.Equal(t, Assigned, conversation.Status())
			assert.Equal(t, found, conversation.FoundMessages())
			assert.Equal(t, added, conversation.AddedMessages())
			assert.Equal(t, updated, conversation.UpdatedMessages())
			assert.Equal(t, deleted, conversation.DeletedMessages())
			assert.Equal(t, 5, conversation.UnreadCount())
			assert.Equal(t, tt.expected, conversation.AgentID())
			assert.Equal(t, "contact-1", conversation.ContactID())
			assert.Equal(t, baseTime, conversation.CreatedAt())
			assert.Equal(t, &updatedAt, conversation.UpdatedAt())
			assert.Equal(t, &finishedAt, conversation.FinishedAt())
			assert.Equal(t, &deletedAt, conversation.DeletedAt())
		})
	}
}

func TestCanAddMessage(t *testing.T) {
	tests := []struct {
		title    string
		status   ConversationStatus
		expected bool
	}{
		{
			title:    "success - pending conversation accepts messages",
			status:   Pending,
			expected: true,
		},
		{
			title:    "success - assigned conversation accepts messages",
			status:   Assigned,
			expected: true,
		},
		{
			title:    "failure - expired conversation rejects messages",
			status:   Expired,
			expected: false,
		},
		{
			title:    "failure - resolved conversation rejects messages",
			status:   Resolved,
			expected: false,
		},
		{
			title:    "success - unknown status accepts messages boundary",
			status:   ConversationStatus(""),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			conversation := newTestConversation(t, tt.status, nil, 0)

			// Act
			canAdd := conversation.CanAddMessage()

			// Assert
			assert.Equal(t, tt.expected, canAdd)
		})
	}
}

func TestAddMessage(t *testing.T) {
	contactID := "contact-1"
	receivedAt := baseTime.Add(30 * time.Minute)

	tests := []struct {
		title              string
		convStatus         ConversationStatus
		existingMessage    bool
		contactID          *string
		expectedConvStatus ConversationStatus
		expectAdded        bool
		expectedMsgStatus  MessageStatus
		expectedFromAgent  bool
	}{
		{
			title:              "success - contact message is delivered and keeps conversation status",
			convStatus:         Pending,
			contactID:          &contactID,
			expectedConvStatus: Pending,
			expectAdded:        true,
			expectedMsgStatus:  Delivered,
			expectedFromAgent:  false,
		},
		{
			title:              "success - agent message is sent and assigns conversation",
			convStatus:         Pending,
			contactID:          nil,
			expectedConvStatus: Assigned,
			expectAdded:        true,
			expectedMsgStatus:  Sent,
			expectedFromAgent:  true,
		},
		{
			title:              "success - skips message already known in conversation",
			convStatus:         Pending,
			existingMessage:    true,
			contactID:          &contactID,
			expectedConvStatus: Pending,
			expectAdded:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			found := map[string]*Message{}
			if tt.existingMessage {
				found["m-1"] = newTestMessage(t, "m-1", Delivered, nil, nil, nil)
			}
			conversation := newTestConversation(t, tt.convStatus, found, 0)

			// Act
			err := conversation.AddMessage("m-1", "new text", tt.contactID, receivedAt)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConvStatus, conversation.Status())
			if tt.expectAdded {
				msg, ok := conversation.AddedMessages()["m-1"]
				require.True(t, ok)
				assert.Equal(t, "new text", msg.Text())
				assert.Equal(t, tt.expectedMsgStatus, msg.Status())
				assert.Equal(t, tt.expectedFromAgent, msg.FromAgent())
				assert.True(t, receivedAt.Equal(msg.CreatedAt()))
				assert.Same(t, conversation.FoundMessages()["m-1"], msg)
				require.NotNil(t, conversation.UpdatedAt())
				assert.True(t, receivedAt.Equal(*conversation.UpdatedAt()))
			} else {
				assert.Empty(t, conversation.AddedMessages())
				assert.Len(t, conversation.FoundMessages(), 1)
				assert.Nil(t, conversation.UpdatedAt())
			}
		})
	}
}

func TestUpdateMessage(t *testing.T) {
	firstEdit := baseTime.Add(time.Hour)
	secondEdit := baseTime.Add(2 * time.Hour)
	firstEditOtherZone := firstEdit.In(time.FixedZone("UTC-5", -5*3600))

	tests := []struct {
		title         string
		foundMsg      *Message
		editedAt      time.Time
		expectedError string
		expectUpdated bool
	}{
		{
			title:         "failure - rejects unknown message",
			foundMsg:      nil,
			editedAt:      firstEdit,
			expectedError: ErrMessageNotFound.Error(),
			expectUpdated: false,
		},
		{
			title:         "success - updates text and timestamp of message without prior edit",
			foundMsg:      newTestMessage(t, "m-1", Sent, nil, nil, nil),
			editedAt:      firstEdit,
			expectUpdated: true,
		},
		{
			title:         "success - updates message with previous edit at different time",
			foundMsg:      newTestMessage(t, "m-1", Sent, &secondEdit, nil, nil),
			editedAt:      firstEdit,
			expectUpdated: true,
		},
		{
			title:         "success - skips edit already applied regardless of zone",
			foundMsg:      newTestMessage(t, "m-1", Sent, &firstEditOtherZone, nil, nil),
			editedAt:      firstEdit,
			expectUpdated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			found := map[string]*Message{}
			if tt.foundMsg != nil {
				found["m-1"] = tt.foundMsg
			}
			conversation := newTestConversation(t, Assigned, found, 0)

			// Act
			err := conversation.UpdateMessage("m-1", "edited text", tt.editedAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrMessageNotFound)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			msg, ok := conversation.UpdatedMessages()["m-1"]
			assert.Equal(t, tt.expectUpdated, ok)
			if tt.expectUpdated {
				assert.Equal(t, "edited text", msg.Text())
				require.NotNil(t, msg.UpdatedAt())
				assert.True(t, tt.editedAt.Equal(*msg.UpdatedAt()))
				assert.Same(t, conversation.FoundMessages()["m-1"], msg)
				require.NotNil(t, conversation.UpdatedAt())
				assert.True(t, tt.editedAt.Equal(*conversation.UpdatedAt()))
			} else {
				assert.Equal(t, "original text", tt.foundMsg.Text())
				assert.Nil(t, conversation.UpdatedAt())
			}
		})
	}
}

func TestDeleteMessage(t *testing.T) {
	firstDelete := baseTime.Add(time.Hour)
	secondDelete := baseTime.Add(2 * time.Hour)

	tests := []struct {
		title         string
		foundMsg      *Message
		deletedAt     time.Time
		expectedError string
		expectDeleted bool
	}{
		{
			title:         "failure - rejects unknown message",
			foundMsg:      nil,
			deletedAt:     firstDelete,
			expectedError: ErrMessageNotFound.Error(),
			expectDeleted: false,
		},
		{
			title:         "success - marks message as deleted",
			foundMsg:      newTestMessage(t, "m-1", Delivered, nil, nil, nil),
			deletedAt:     firstDelete,
			expectDeleted: true,
		},
		{
			title:         "success - skips already deleted message",
			foundMsg:      newTestMessage(t, "m-1", Deleted, nil, &firstDelete, nil),
			deletedAt:     secondDelete,
			expectDeleted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			found := map[string]*Message{}
			if tt.foundMsg != nil {
				found["m-1"] = tt.foundMsg
			}
			conversation := newTestConversation(t, Assigned, found, 0)

			// Act
			err := conversation.DeleteMessage("m-1", tt.deletedAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrMessageNotFound)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			msg, ok := conversation.DeletedMessages()["m-1"]
			assert.Equal(t, tt.expectDeleted, ok)
			if tt.expectDeleted {
				assert.Equal(t, Deleted, msg.Status())
				require.NotNil(t, msg.DeletedAt())
				assert.True(t, tt.deletedAt.Equal(*msg.DeletedAt()))
				assert.Same(t, conversation.FoundMessages()["m-1"], msg)
				require.NotNil(t, conversation.UpdatedAt())
				assert.True(t, tt.deletedAt.Equal(*conversation.UpdatedAt()))
			} else {
				require.NotNil(t, tt.foundMsg.DeletedAt())
				assert.True(t, firstDelete.Equal(*tt.foundMsg.DeletedAt()))
				assert.Nil(t, conversation.UpdatedAt())
			}
		})
	}
}

func TestMarkMessageRead(t *testing.T) {
	firstRead := baseTime.Add(time.Hour)
	firstReadOtherZone := firstRead.In(time.FixedZone("UTC-5", -5*3600))

	tests := []struct {
		title         string
		foundMsg      *Message
		readAt        time.Time
		expectedError string
		expectUpdated bool
	}{
		{
			title:         "failure - rejects unknown message",
			foundMsg:      nil,
			readAt:        firstRead,
			expectedError: ErrMessageNotFound.Error(),
			expectUpdated: false,
		},
		{
			title:         "success - marks message as read and clears unread count",
			foundMsg:      newTestMessage(t, "m-1", Delivered, nil, nil, nil),
			readAt:        firstRead,
			expectUpdated: true,
		},
		{
			title:         "success - skips read receipt already applied regardless of zone",
			foundMsg:      newTestMessage(t, "m-1", Read, nil, nil, &firstReadOtherZone),
			readAt:        firstRead,
			expectUpdated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			found := map[string]*Message{}
			if tt.foundMsg != nil {
				found["m-1"] = tt.foundMsg
			}
			conversation := newTestConversation(t, Assigned, found, 3)

			// Act
			err := conversation.MarkMessageRead("m-1", tt.readAt)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrMessageNotFound)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			msg, ok := conversation.UpdatedMessages()["m-1"]
			assert.Equal(t, tt.expectUpdated, ok)
			if tt.expectUpdated {
				assert.Equal(t, Read, msg.Status())
				require.NotNil(t, msg.ReadAt())
				assert.True(t, tt.readAt.Equal(*msg.ReadAt()))
				assert.Same(t, conversation.FoundMessages()["m-1"], msg)
				assert.Equal(t, 0, conversation.UnreadCount())
				require.NotNil(t, conversation.UpdatedAt())
				assert.True(t, tt.readAt.Equal(*conversation.UpdatedAt()))
			} else {
				require.NotNil(t, tt.foundMsg.ReadAt())
				assert.True(t, firstRead.Equal(*tt.foundMsg.ReadAt()))
				assert.Equal(t, 3, conversation.UnreadCount())
				assert.Nil(t, conversation.UpdatedAt())
			}
		})
	}
}
