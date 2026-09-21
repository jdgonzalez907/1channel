package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	id := uuid.NewV7()
	externalID := "wa-message-1"
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	updatedAt := baseTime.Add(time.Hour)
	deletedAt := baseTime.Add(2 * time.Hour)
	readAt := baseTime.Add(3 * time.Hour)

	tests := []struct {
		title      string
		externalID *string
		text       string
		status     MessageStatus
		agentID    *uuid.UUID
		contactID  *uuid.UUID
		updatedAt  *time.Time
		deletedAt  *time.Time
		readAt     *time.Time
	}{
		{
			title:      "success - builds message with all fields populated",
			externalID: &externalID,
			text:       "hello",
			status:     Read,
			agentID:    &agentID,
			contactID:  &contactID,
			updatedAt:  &updatedAt,
			deletedAt:  &deletedAt,
			readAt:     &readAt,
		},
		{
			title:      "success - builds message with nil optional fields boundary",
			externalID: nil,
			text:       "",
			status:     Registered,
			agentID:    nil,
			contactID:  nil,
			updatedAt:  nil,
			deletedAt:  nil,
			readAt:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			message, err := NewMessage(id, tt.externalID, tt.text, tt.status, tt.agentID, tt.contactID, baseTime, tt.updatedAt, tt.deletedAt, tt.readAt)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, message)
			assert.Equal(t, id, message.ID())
			assert.Equal(t, tt.externalID, message.ExternalID())
			assert.Equal(t, tt.text, message.Text())
			assert.Equal(t, tt.status, message.Status())
			assert.Equal(t, tt.agentID, message.AgentID())
			assert.Equal(t, tt.contactID, message.ContactID())
			assert.Equal(t, baseTime, message.CreatedAt())
			assert.Equal(t, tt.updatedAt, message.UpdatedAt())
			assert.Equal(t, tt.deletedAt, message.DeletedAt())
			assert.Equal(t, tt.readAt, message.ReadAt())
		})
	}
}

func TestMarkAsSent(t *testing.T) {
	at := baseTime.Add(time.Hour)

	tests := []struct {
		title          string
		currentStatus  MessageStatus
		expected       bool
		expectedStatus MessageStatus
	}{
		{
			title:          "success - marks registered message as sent",
			currentStatus:  Registered,
			expected:       true,
			expectedStatus: Sent,
		},
		{
			title:          "success - marks failed message as sent (retry)",
			currentStatus:  Failed,
			expected:       true,
			expectedStatus: Sent,
		},
		{
			title:          "failure - keeps sent message untouched",
			currentStatus:  Sent,
			expected:       false,
			expectedStatus: Sent,
		},
		{
			title:          "failure - keeps delivered message untouched",
			currentStatus:  Delivered,
			expected:       false,
			expectedStatus: Delivered,
		},
		{
			title:          "failure - keeps read message untouched",
			currentStatus:  Read,
			expected:       false,
			expectedStatus: Read,
		},
		{
			title:          "failure - keeps deleted message untouched",
			currentStatus:  Deleted,
			expected:       false,
			expectedStatus: Deleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			agentID := uuid.NewV7()
			message := newTestAgentMessage(t, agentID, tt.currentStatus)

			// Act
			changed := message.MarkAsSent(at)

			// Assert
			assert.Equal(t, tt.expected, changed)
			assert.Equal(t, tt.expectedStatus, message.Status())
		})
	}
}

func TestMarkAsDelivered(t *testing.T) {
	at := baseTime.Add(time.Hour)

	tests := []struct {
		title          string
		currentStatus  MessageStatus
		expected       bool
		expectedStatus MessageStatus
	}{
		{
			title:          "success - marks sent message as delivered",
			currentStatus:  Sent,
			expected:       true,
			expectedStatus: Delivered,
		},
		{
			title:          "failure - keeps registered message untouched",
			currentStatus:  Registered,
			expected:       false,
			expectedStatus: Registered,
		},
		{
			title:          "failure - keeps delivered message untouched",
			currentStatus:  Delivered,
			expected:       false,
			expectedStatus: Delivered,
		},
		{
			title:          "failure - keeps read message untouched",
			currentStatus:  Read,
			expected:       false,
			expectedStatus: Read,
		},
		{
			title:          "failure - keeps deleted message untouched",
			currentStatus:  Deleted,
			expected:       false,
			expectedStatus: Deleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			agentID := uuid.NewV7()
			message := newTestAgentMessage(t, agentID, tt.currentStatus)

			// Act
			changed := message.MarkAsDelivered(at)

			// Assert
			assert.Equal(t, tt.expected, changed)
			assert.Equal(t, tt.expectedStatus, message.Status())
		})
	}
}

func TestMarkAsRead(t *testing.T) {
	at := baseTime.Add(time.Hour)

	tests := []struct {
		title          string
		status         MessageStatus
		expected       bool
		expectedStatus MessageStatus
	}{
		{
			title:          "success - marks delivered message as read",
			status:         Delivered,
			expected:       true,
			expectedStatus: Read,
		},
		{
			title:          "success - marks registered message as read jumping forward",
			status:         Registered,
			expected:       true,
			expectedStatus: Read,
		},
		{
			title:          "failure - keeps deleted message untouched",
			status:         Deleted,
			expected:       false,
			expectedStatus: Deleted,
		},
		{
			title:          "failure - keeps failed message untouched",
			status:         Failed,
			expected:       false,
			expectedStatus: Failed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			contactID := uuid.NewV7()
			message := newTestContactMessage(t, contactID, tt.status, "wa-message-1")

			// Act
			changed := message.MarkAsRead(at)

			// Assert
			assert.Equal(t, tt.expected, changed)
			assert.Equal(t, tt.expectedStatus, message.Status())
			if tt.expected {
				require.NotNil(t, message.ReadAt())
				assert.True(t, at.Equal(*message.ReadAt()))
			} else {
				assert.Nil(t, message.ReadAt())
			}
		})
	}
}

func TestMarkAsFailed(t *testing.T) {
	at := baseTime.Add(time.Hour)

	tests := []struct {
		title          string
		currentStatus  MessageStatus
		expected       bool
		expectedStatus MessageStatus
	}{
		{
			title:          "success - marks registered message as failed",
			currentStatus:  Registered,
			expected:       true,
			expectedStatus: Failed,
		},
		{
			title:          "success - marks sent message as failed",
			currentStatus:  Sent,
			expected:       true,
			expectedStatus: Failed,
		},
		{
			title:          "success - marks delivered message as failed",
			currentStatus:  Delivered,
			expected:       true,
			expectedStatus: Failed,
		},
		{
			title:          "success - marks read message as failed",
			currentStatus:  Read,
			expected:       true,
			expectedStatus: Failed,
		},
		{
			title:          "failure - keeps failed message untouched",
			currentStatus:  Failed,
			expected:       false,
			expectedStatus: Failed,
		},
		{
			title:          "failure - keeps deleted message untouched",
			currentStatus:  Deleted,
			expected:       false,
			expectedStatus: Deleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			agentID := uuid.NewV7()
			message := newTestAgentMessage(t, agentID, tt.currentStatus)

			// Act
			changed := message.MarkAsFailed(at)

			// Assert
			assert.Equal(t, tt.expected, changed)
			assert.Equal(t, tt.expectedStatus, message.Status())
		})
	}
}

func TestUpdateText(t *testing.T) {
	// Arrange
	editedAt := baseTime.Add(time.Hour)
	agentID := uuid.NewV7()
	message := newTestAgentMessage(t, agentID, Registered)

	// Act
	message.UpdateText("edited text", editedAt)

	// Assert
	assert.Equal(t, "edited text", message.Text())
	require.NotNil(t, message.UpdatedAt())
	assert.True(t, editedAt.Equal(*message.UpdatedAt()))
}

func TestMarkAsDeleted(t *testing.T) {
	firstDelete := baseTime.Add(time.Hour)
	secondDelete := baseTime.Add(2 * time.Hour)

	tests := []struct {
		title        string
		status       MessageStatus
		deletedAt    *time.Time
		at           time.Time
		expected     bool
		expectedTime *time.Time
	}{
		{
			title:        "success - deletes delivered message",
			status:       Delivered,
			deletedAt:    nil,
			at:           firstDelete,
			expected:     true,
			expectedTime: &firstDelete,
		},
		{
			title:        "success - deletes already read message",
			status:       Read,
			deletedAt:    nil,
			at:           firstDelete,
			expected:     true,
			expectedTime: &firstDelete,
		},
		{
			title:        "failure - skips already deleted message keeping first timestamp",
			status:       Deleted,
			deletedAt:    &firstDelete,
			at:           secondDelete,
			expected:     false,
			expectedTime: &firstDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			agentID := uuid.NewV7()
			message := newTestAgentMessage(t, agentID, tt.status)
			if tt.deletedAt != nil {
				message.deletedAt = tt.deletedAt
			}

			// Act
			changed := message.MarkAsDeleted(tt.at)

			// Assert
			assert.Equal(t, tt.expected, changed)
			assert.Equal(t, Deleted, message.Status())
			require.NotNil(t, message.DeletedAt())
			assert.True(t, tt.expectedTime.Equal(*message.DeletedAt()))
		})
	}
}

func TestAssignExternalID(t *testing.T) {
	existingID := "wa-existing"
	newID := "wa-new"

	tests := []struct {
		title      string
		currentID  *string
		inputID    string
		expected   bool
		expectedID string
	}{
		{
			title:      "success - assigns channel id when message has none",
			currentID:  nil,
			inputID:    newID,
			expected:   true,
			expectedID: newID,
		},
		{
			title:      "failure - does not overwrite an already assigned channel id",
			currentID:  &existingID,
			inputID:    newID,
			expected:   false,
			expectedID: existingID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			agentID := uuid.NewV7()
			message := newTestAgentMessage(t, agentID, Registered)
			message.externalID = tt.currentID

			// Act
			assigned := message.AssignExternalID(tt.inputID)

			// Assert
			assert.Equal(t, tt.expected, assigned)
			require.NotNil(t, message.ExternalID())
			assert.Equal(t, tt.expectedID, *message.ExternalID())
		})
	}
}
