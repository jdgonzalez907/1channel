package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MessageType
		wantErr bool
	}{
		{name: "text", input: "text", want: MessageTypeText, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
		{name: "empty", input: "", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMessageType(tt.input)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrMessageTypeInvalid)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMessageType_String(t *testing.T) {
	assert.Equal(t, "text", MessageTypeText.String())
}
