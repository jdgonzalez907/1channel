package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConversationStatus(t *testing.T) {
	tests := []struct {
		title    string
		value    string
		expected ConversationStatus
	}{
		{
			title:    "success - converts known status value",
			value:    "assigned",
			expected: Assigned,
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

func TestConversationStatusValue(t *testing.T) {
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
