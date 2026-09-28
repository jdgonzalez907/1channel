package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageStatus(t *testing.T) {
	// Arrange
	tests := []struct {
		name    string
		input   string
		want    MessageStatus
		wantErr bool
	}{
		{name: "sent", input: "sent", want: MessageStatusSent, wantErr: false},
		{name: "read", input: "read", want: MessageStatusRead, wantErr: false},
		{name: "deleted", input: "deleted", want: MessageStatusDeleted, wantErr: false},
		{name: "failed", input: "failed", want: MessageStatusFailed, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
		{name: "empty", input: "", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got, err := NewMessageStatus(tt.input)

			// Assert
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrMessageStatusInvalid)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMessageStatus_String(t *testing.T) {
	// Arrange
	tests := []struct {
		name   string
		status MessageStatus
		want   string
	}{
		{name: "sent", status: MessageStatusSent, want: "sent"},
		{name: "read", status: MessageStatusRead, want: "read"},
		{name: "deleted", status: MessageStatusDeleted, want: "deleted"},
		{name: "failed", status: MessageStatusFailed, want: "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := tt.status.String()

			// Assert
			assert.Equal(t, tt.want, got)
		})
	}
}
