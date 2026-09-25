package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"uuid"
)

func TestNewMessage(t *testing.T) {
	validID := uuid.NewV7()
	now := time.Now()
	text := "hello"

	tests := []struct {
		name      string
		id        uuid.UUID
		status    MessageStatus
		msgType   MessageType
		text      *string
		agentID   *uuid.UUID
		contactID *uuid.UUID
		sentAt    time.Time
		wantErr   error
	}{
		{
			name:    "valid message",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    &text,
			sentAt:  now,
			wantErr: nil,
		},
		{
			name:    "nil uuid",
			id:      uuid.Nil(),
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    &text,
			sentAt:  now,
			wantErr: ErrMessageInvalidID,
		},
		{
			name:    "empty text for text type",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    nil,
			sentAt:  now,
			wantErr: ErrMessageEmptyText,
		},
		{
			name:    "text too long",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    longText(1001),
			sentAt:  now,
			wantErr: ErrMessageTextTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewMessage(tt.id, tt.status, tt.msgType, tt.text, tt.agentID, tt.contactID, tt.sentAt, nil, nil, nil)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, msg)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, msg)
		})
	}
}

func TestMessage_Getters(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()
	text := "hello"
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()

	msg, err := NewMessage(id, MessageStatusSent, MessageTypeText, &text, &agentID, &contactID, now, nil, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, id, msg.ID())
	assert.Equal(t, MessageStatusSent, msg.Status())
	assert.Equal(t, MessageTypeText, msg.Type())
	assert.Equal(t, &text, msg.Text())
	assert.Equal(t, &agentID, msg.AgentID())
	assert.Equal(t, &contactID, msg.ContactID())
	assert.Equal(t, now, msg.SentAt())
	assert.Nil(t, msg.ReadAt())
	assert.Nil(t, msg.EditedAt())
	assert.Nil(t, msg.DeletedAt())
}

func longText(n int) *string {
	s := make([]rune, n)
	for i := range s {
		s[i] = 'a'
	}
	str := string(s)
	return &str
}
