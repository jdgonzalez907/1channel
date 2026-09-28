package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	validID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()

	tests := []struct {
		name       string
		id         uuid.UUID
		status     MessageStatus
		msgType    MessageType
		text       *string
		agentID    *uuid.UUID
		contactID  *uuid.UUID
		externalID *string
		sentAt     time.Time
		wantErr    error
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
			name:       "valid message with external id",
			id:         validID,
			status:     MessageStatusSent,
			msgType:    MessageTypeText,
			text:       &text,
			contactID:  &contactID,
			externalID: strPtr("wa-msg-001"),
			sentAt:     now,
			wantErr:    nil,
		},
		{
			name:       "empty external id",
			id:         validID,
			status:     MessageStatusSent,
			msgType:    MessageTypeText,
			text:       &text,
			contactID:  &contactID,
			externalID: strPtr(""),
			sentAt:     now,
			wantErr:    ErrMessageExternalIDInvalid,
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
			msg, err := NewMessage(tt.id, tt.status, tt.msgType, tt.text, tt.agentID, tt.contactID, tt.externalID, tt.sentAt, nil, nil, nil)

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

	msg, err := NewMessage(id, MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)

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
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, time.Now(), nil, nil, nil)

	// Act
	err := msg.AssignExternalID("wa-msg-001")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "wa-msg-001", *msg.ExternalID())
}

func TestMessage_AssignExternalID_AlreadyAssigned(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, time.Now(), nil, nil, nil)
	_ = msg.AssignExternalID("wa-msg-001")

	// Act
	err := msg.AssignExternalID("wa-msg-002")

	// Assert
	assert.ErrorIs(t, err, ErrMessageExternalIDAlreadySet)
	assert.Equal(t, "wa-msg-001", *msg.ExternalID())
}

func TestMessage_AssignExternalID_EmptyValue(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, time.Now(), nil, nil, nil)

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
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)

	readAt := now.Add(time.Hour)

	// Act
	msg.MarkAsRead(readAt)

	// Assert
	assert.Equal(t, MessageStatusRead, msg.Status())
	assert.Equal(t, &readAt, msg.ReadAt())
}

func TestMessage_EditText(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	now := time.Now()
	editedAt := now.Add(time.Hour)
	deletedAt := now.Add(2 * time.Hour)

	tests := []struct {
		name     string
		msgFunc  func() *Message
		newText  string
		at       time.Time
		wantErr  error
		wantText string
		wantEdit bool
	}{
		{
			name: "success",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			newText:  "edited",
			at:       editedAt,
			wantErr:  nil,
			wantText: "edited",
			wantEdit: true,
		},
		{
			name: "empty text",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			newText: "",
			at:      editedAt,
			wantErr: ErrMessageEmptyText,
		},
		{
			name: "text too long",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			newText: *longText(MaxTextLength + 1),
			at:      editedAt,
			wantErr: ErrMessageTextTooLong,
		},
		{
			name: "message without text",
			msgFunc: func() *Message {
				return &Message{id: uuid.NewV7(), status: MessageStatusSent, msgType: MessageTypeText, agentID: &agentID, sentAt: now}
			},
			newText: "edited",
			at:      editedAt,
			wantErr: ErrMessageNotText,
		},
		{
			name: "failed message",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusFailed, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			newText: "edited",
			at:      editedAt,
			wantErr: ErrMessageFailed,
		},
		{
			name: "deleted message with later timestamp",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusDeleted, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, &deletedAt)
				return msg
			},
			newText: "edited",
			at:      deletedAt,
			wantErr: ErrMessageAlreadyDeleted,
		},
		{
			name: "deleted message with earlier timestamp",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusDeleted, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, &deletedAt)
				return msg
			},
			newText:  "edited",
			at:       now,
			wantErr:  nil,
			wantText: "edited",
			wantEdit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.msgFunc()

			// Act
			err := msg.EditText(tt.newText, tt.at)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantText, *msg.Text())

			if tt.wantEdit {
				assert.Equal(t, tt.at, *msg.EditedAt())
			}
		})
	}
}

func TestMessage_Delete(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	now := time.Now()
	later := now.Add(time.Hour)
	deletedAt := now.Add(-time.Hour)

	tests := []struct {
		name    string
		msgFunc func() *Message
		at      time.Time
		wantErr error
	}{
		{
			name: "success",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			at:      later,
			wantErr: nil,
		},
		{
			name: "already deleted with later timestamp",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusDeleted, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, &deletedAt)
				return msg
			},
			at:      now,
			wantErr: ErrMessageAlreadyDeleted,
		},
		{
			name: "already deleted with earlier timestamp",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusDeleted, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, &deletedAt)
				return msg
			},
			at:      deletedAt.Add(-time.Hour),
			wantErr: ErrMessageAlreadyDeleted,
		},
		{
			name: "failed message",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusFailed, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			at:      now,
			wantErr: ErrMessageFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.msgFunc()

			// Act
			err := msg.Delete(tt.at)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, MessageStatusDeleted, msg.Status())
			assert.Equal(t, tt.at, *msg.DeletedAt())
		})
	}
}

func TestMessage_MarkAsFailed(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name    string
		msgFunc func() *Message
		wantErr error
	}{
		{
			name: "success",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			wantErr: nil,
		},
		{
			name: "idempotent",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusFailed, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			wantErr: nil,
		},
		{
			name: "message not from agent",
			msgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), nil, &contactID, nil, now, nil, nil, nil)
				return msg
			},
			wantErr: ErrMessageNotFromAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.msgFunc()

			// Act
			err := msg.MarkAsFailed()

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, MessageStatusFailed, msg.Status())
		})
	}
}

func TestMessage_EditText_DiscardsStaleEdit(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	now := time.Now()
	newestAt := now.Add(time.Hour)

	tests := []struct {
		name    string
		newText string
		at      time.Time
	}{
		{name: "earlier timestamp", newText: "stale", at: now},
		{name: "equal timestamp", newText: "equal", at: newestAt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, strPtr("hello"), &agentID, nil, nil, now, nil, nil, nil)
			_ = msg.EditText("newest", newestAt)

			// Act
			err := msg.EditText(tt.newText, tt.at)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, "newest", *msg.Text())
			assert.Equal(t, &newestAt, msg.EditedAt())
		})
	}
}

func TestRehydrateMessage(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()
	text := "hello"
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	externalID := "wa-msg-001"
	readAt := now.Add(time.Minute)
	editedAt := now.Add(2 * time.Minute)
	deletedAt := now.Add(3 * time.Minute)

	tests := []struct {
		name       string
		id         uuid.UUID
		status     MessageStatus
		msgType    MessageType
		text       *string
		agentID    *uuid.UUID
		contactID  *uuid.UUID
		externalID *string
	}{
		{
			name:       "hydrates an agent message",
			id:         id,
			status:     MessageStatusRead,
			msgType:    MessageTypeText,
			text:       &text,
			agentID:    &agentID,
			externalID: &externalID,
		},
		{
			name:      "maps an invalid state directly",
			id:        id,
			status:    MessageStatusDeleted,
			msgType:   MessageTypeText,
			text:      nil,
			agentID:   &agentID,
			contactID: &contactID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			msg := RehydrateMessage(tt.id, tt.status, tt.msgType, tt.text, tt.agentID, tt.contactID, tt.externalID, now, &readAt, &editedAt, &deletedAt)

			// Assert
			assert.NotNil(t, msg)
			assert.Equal(t, tt.id, msg.ID())
			assert.Equal(t, tt.status, msg.Status())
			assert.Equal(t, tt.msgType, msg.Type())
			assert.Equal(t, tt.text, msg.Text())
			assert.Equal(t, tt.agentID, msg.AgentID())
			assert.Equal(t, tt.contactID, msg.ContactID())
			assert.Equal(t, tt.externalID, msg.ExternalID())
			assert.Equal(t, now, msg.SentAt())
			assert.Equal(t, &readAt, msg.ReadAt())
			assert.Equal(t, &editedAt, msg.EditedAt())
			assert.Equal(t, &deletedAt, msg.DeletedAt())
		})
	}
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
