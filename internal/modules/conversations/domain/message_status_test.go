package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MessageStatus
		wantErr bool
	}{
		{name: "sent", input: "sent", want: MessageStatusSent, wantErr: false},
		{name: "read", input: "read", want: MessageStatusRead, wantErr: false},
		{name: "deleted", input: "deleted", want: MessageStatusDeleted, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
		{name: "empty", input: "", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMessageStatus(tt.input)

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
	assert.Equal(t, "sent", MessageStatusSent.String())
	assert.Equal(t, "read", MessageStatusRead.String())
}
