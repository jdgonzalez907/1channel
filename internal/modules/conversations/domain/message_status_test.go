package domain

import (
	"testing"

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

func TestMessageStatusValue(t *testing.T) {
	tests := []struct {
		title    string
		status   MessageStatus
		expected string
	}{
		{
			title:    "success - returns string of registered status",
			status:   Registered,
			expected: "registered",
		},
		{
			title:    "success - returns string of deleted status",
			status:   Deleted,
			expected: "deleted",
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
