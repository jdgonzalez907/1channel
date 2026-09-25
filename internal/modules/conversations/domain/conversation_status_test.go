package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConversationStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ConversationStatus
		wantErr bool
	}{
		{name: "pending", input: "pending", want: ConversationStatusPending, wantErr: false},
		{name: "assigned", input: "assigned", want: ConversationStatusAssigned, wantErr: false},
		{name: "expired", input: "expired", want: ConversationStatusExpired, wantErr: false},
		{name: "resolved", input: "resolved", want: ConversationStatusResolved, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
		{name: "empty", input: "", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConversationStatus(tt.input)

			if tt.wantErr {
				assert.ErrorIs(t, err, ErrConversationStatusInvalid)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConversationStatus_String(t *testing.T) {
	assert.Equal(t, "pending", ConversationStatusPending.String())
	assert.Equal(t, "assigned", ConversationStatusAssigned.String())
}
