package conversations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessageStatus(t *testing.T) {
	tests := []struct {
		title    string
		value    string
		expected MessageStatus
	}{
		{
			title:    "success - converts known status value",
			value:    "delivered",
			expected: Delivered,
		},
		{
			title:    "success - converts empty value boundary",
			value:    "",
			expected: MessageStatus(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			status, err := NewMessageStatus(tt.value)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestMessageValue(t *testing.T) {
	tests := []struct {
		title    string
		status   MessageStatus
		expected string
	}{
		{
			title:    "success - returns string of sent status",
			status:   Sent,
			expected: "sent",
		},
		{
			title:    "success - returns string of read status",
			status:   Read,
			expected: "read",
		},
		{
			title:    "success - returns empty string for empty status boundary",
			status:   MessageStatus(""),
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

func TestNewMessage(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)
	readAt := time.Date(2026, 1, 4, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		title     string
		id        string
		text      string
		status    MessageStatus
		fromAgent bool
		createdAt time.Time
		updatedAt *time.Time
		deletedAt *time.Time
		readAt    *time.Time
	}{
		{
			title:     "success - builds message with all fields populated",
			id:        "msg-1",
			text:      "hello",
			status:    Read,
			fromAgent: true,
			createdAt: createdAt,
			updatedAt: &updatedAt,
			deletedAt: &deletedAt,
			readAt:    &readAt,
		},
		{
			title:     "success - builds message with nil optional times boundary",
			id:        "msg-2",
			text:      "",
			status:    Sent,
			fromAgent: false,
			createdAt: createdAt,
			updatedAt: nil,
			deletedAt: nil,
			readAt:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			msg, err := NewMessage(tt.id, tt.text, tt.status, tt.fromAgent, tt.createdAt, tt.updatedAt, tt.deletedAt, tt.readAt)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, msg)
			assert.Equal(t, tt.id, msg.ID())
			assert.Equal(t, tt.text, msg.Text())
			assert.Equal(t, tt.status, msg.Status())
			assert.Equal(t, tt.fromAgent, msg.FromAgent())
			assert.Equal(t, tt.createdAt, msg.CreatedAt())
			assert.Equal(t, tt.updatedAt, msg.UpdatedAt())
			assert.Equal(t, tt.deletedAt, msg.DeletedAt())
			assert.Equal(t, tt.readAt, msg.ReadAt())
		})
	}
}
