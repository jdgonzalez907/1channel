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
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()

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
			name:    "valid message from agent",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    &text,
			agentID: &agentID,
			sentAt:  now,
			wantErr: nil,
		},
		{
			name:      "valid message from contact",
			id:        validID,
			status:    MessageStatusSent,
			msgType:   MessageTypeText,
			text:      &text,
			contactID: &contactID,
			sentAt:    now,
			wantErr:   nil,
		},
		{
			name:    "nil uuid",
			id:      uuid.Nil(),
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    &text,
			agentID: &agentID,
			sentAt:  now,
			wantErr: ErrMessageInvalidID,
		},
		{
			name:    "empty text for text type",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    nil,
			agentID: &agentID,
			sentAt:  now,
			wantErr: ErrMessageEmptyText,
		},
		{
			name:    "blank text for text type",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    strPtr(""),
			agentID: &agentID,
			sentAt:  now,
			wantErr: ErrMessageEmptyText,
		},
		{
			name:    "text too long",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    longText(1001),
			agentID: &agentID,
			sentAt:  now,
			wantErr: ErrMessageTextTooLong,
		},
		{
			name:    "no owner",
			id:      validID,
			status:  MessageStatusSent,
			msgType: MessageTypeText,
			text:    &text,
			sentAt:  now,
			wantErr: ErrMessageInvalidOwner,
		},
		{
			name:      "two owners",
			id:        validID,
			status:    MessageStatusSent,
			msgType:   MessageTypeText,
			text:      &text,
			agentID:   &agentID,
			contactID: &contactID,
			sentAt:    now,
			wantErr:   ErrMessageInvalidOwner,
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
	// Arrange
	id := uuid.NewV7()
	now := time.Now()
	text := "hello"
	agentID := uuid.NewV7()

	msg, err := NewMessage(id, MessageStatusSent, MessageTypeText, &text, &agentID, nil, now, nil, nil, nil)

	// Act
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, id, msg.ID())
	assert.Equal(t, MessageStatusSent, msg.Status())
	assert.Equal(t, MessageTypeText, msg.Type())
	assert.Equal(t, &text, msg.Text())
	assert.Equal(t, &agentID, msg.AgentID())
	assert.Nil(t, msg.ContactID())
	assert.Equal(t, now, msg.SentAt())
	assert.Nil(t, msg.ReadAt())
	assert.Nil(t, msg.EditedAt())
	assert.Nil(t, msg.DeletedAt())
	assert.Nil(t, msg.ExternalID())
}

func TestMessage_AssignExternalID_Success(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, time.Now(), nil, nil, nil)

	// Act
	err := msg.AssignExternalID("wa-msg-001")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "wa-msg-001", *msg.ExternalID())
}

func TestMessage_AssignExternalID_AlreadyAssigned(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, time.Now(), nil, nil, nil)
	msg.AssignExternalID("wa-msg-001")

	// Act
	err := msg.AssignExternalID("wa-msg-002")

	// Assert
	assert.ErrorIs(t, err, ErrMessageExternalIDAlreadySet)
	assert.Equal(t, "wa-msg-001", *msg.ExternalID())
}

func TestMessage_AssignExternalID_EmptyValue(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, time.Now(), nil, nil, nil)

	// Act
	err := msg.AssignExternalID("")

	// Assert
	assert.ErrorIs(t, err, ErrMessageExternalIDInvalid)
	assert.Nil(t, msg.ExternalID())
}

func TestMessage_MarkAsRead(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	now := time.Now()
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, now, nil, nil, nil)

	readAt := now.Add(time.Hour)

	// Act
	msg.MarkAsRead(readAt)

	// Assert
	assert.Equal(t, MessageStatusRead, msg.Status())
	assert.Equal(t, &readAt, msg.ReadAt())
}

func strPtr(s string) *string { return &s }

func longText(n int) *string {
	s := make([]rune, n)
	for i := range s {
		s[i] = 'a'
	}
	str := string(s)
	return &str
}
