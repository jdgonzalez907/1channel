package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"uuid"
)

func TestNewConversation(t *testing.T) {
	validID := uuid.NewV7()
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, nil, now, nil, nil, nil)
	messages := []Message{*msg}

	tests := []struct {
		name      string
		id        uuid.UUID
		status    ConversationStatus
		messages  []Message
		agentID   *uuid.UUID
		contactID *uuid.UUID
		createdAt time.Time
		wantErr   error
	}{
		{
			name:      "valid with agent",
			id:        validID,
			status:    ConversationStatusAssigned,
			messages:  messages,
			agentID:   &agentID,
			createdAt: now,
			wantErr:   nil,
		},
		{
			name:      "valid with contact",
			id:        validID,
			status:    ConversationStatusPending,
			messages:  messages,
			contactID: &contactID,
			createdAt: now,
			wantErr:   nil,
		},
		{
			name:      "valid with both",
			id:        validID,
			status:    ConversationStatusAssigned,
			messages:  messages,
			agentID:   &agentID,
			contactID: &contactID,
			createdAt: now,
			wantErr:   nil,
		},
		{
			name:      "nil uuid",
			id:        uuid.Nil(),
			status:    ConversationStatusPending,
			messages:  messages,
			contactID: &contactID,
			createdAt: now,
			wantErr:   ErrConversationInvalidID,
		},
		{
			name:      "missing contact and agent",
			id:        validID,
			status:    ConversationStatusPending,
			messages:  messages,
			createdAt: now,
			wantErr:   ErrConversationMissingContactAndAgent,
		},
		{
			name:      "empty messages",
			id:        validID,
			status:    ConversationStatusPending,
			messages:  []Message{},
			contactID: &contactID,
			createdAt: now,
			wantErr:   ErrConversationEmptyMessages,
		},
		{
			name:      "nil messages",
			id:        validID,
			status:    ConversationStatusPending,
			messages:  nil,
			contactID: &contactID,
			createdAt: now,
			wantErr:   ErrConversationEmptyMessages,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			conv, err := NewConversation(tt.id, tt.status, tt.messages, tt.agentID, tt.contactID, tt.createdAt, nil, nil)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, conv)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, conv)
		})
	}
}

func TestConversation_Getters(t *testing.T) {
	// Arrange
	id := uuid.NewV7()
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, nil, now, nil, nil, nil)

	// Act
	conv, err := NewConversation(id, ConversationStatusAssigned, []Message{*msg}, &agentID, &contactID, now, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, id, conv.ID())
	assert.Equal(t, ConversationStatusAssigned, conv.Status())
	assert.Equal(t, &agentID, conv.AgentID())
	assert.Equal(t, &contactID, conv.ContactID())
	assert.Equal(t, now, conv.CreatedAt())
	assert.Nil(t, conv.UpdatedAt())
	assert.Nil(t, conv.FinishedAt())
	assert.Len(t, conv.Messages(), 1)
}

func TestConversation_Messages(t *testing.T) {
	// Arrange
	convID := uuid.NewV7()
	agentID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	msg1, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, nil, now, nil, nil, nil)
	msg2, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, nil, now, nil, nil, nil)

	// Act
	conv, err := NewConversation(convID, ConversationStatusAssigned, []Message{*msg1, *msg2}, &agentID, nil, now, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conv.Messages(), 2)
}

func TestConversation_ReceiveContactMessage_Success(t *testing.T) {
	// Arrange
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusPending, []Message{*existingMsg}, nil, &contactID, now, nil, nil)

	newMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	newMsg.AssignExternalID("wa-002")

	// Act
	err := conv.ReceiveContactMessage(contactID, *newMsg, now)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conv.Messages(), 2)
}

func TestConversation_ReceiveContactMessage_ContactNotOwner(t *testing.T) {
	// Arrange
	contactID := uuid.NewV7()
	otherContact := uuid.NewV7()
	now := time.Now()
	text := "hello"
	existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusPending, []Message{*existingMsg}, nil, &contactID, now, nil, nil)

	newMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &otherContact, now, nil, nil, nil)

	// Act
	err := conv.ReceiveContactMessage(otherContact, *newMsg, now)

	// Assert
	assert.ErrorIs(t, err, ErrConversationContactNotOwner)
	assert.Len(t, conv.Messages(), 1)
}

func TestConversation_ReceiveContactMessage_ConversationHasNoContact(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, now, nil, nil, nil)
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusAssigned, []Message{*existingMsg}, &agentID, nil, now, nil, nil)

	newMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)

	// Act
	err := conv.ReceiveContactMessage(contactID, *newMsg, now)

	// Assert
	assert.ErrorIs(t, err, ErrConversationHasNoContact)
}

func TestConversation_ReceiveContactMessage_DuplicateMessage(t *testing.T) {
	// Arrange
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	existingMsg.AssignExternalID("wa-001")
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusPending, []Message{*existingMsg}, nil, &contactID, now, nil, nil)

	dupMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	dupMsg.AssignExternalID("wa-001")

	// Act
	err := conv.ReceiveContactMessage(contactID, *dupMsg, now)

	// Assert
	assert.ErrorIs(t, err, ErrConversationDuplicateMessage)
	assert.Len(t, conv.Messages(), 1)
}

func TestConversation_ReceiveContactMessage_WithoutExternalID(t *testing.T) {
	// Arrange
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusPending, []Message{*existingMsg}, nil, &contactID, now, nil, nil)

	newMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)

	// Act
	err := conv.ReceiveContactMessage(contactID, *newMsg, now)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conv.Messages(), 2)
}

func TestConversation_ReceiveContactMessage_UpdatesTimestamp(t *testing.T) {
	// Arrange
	contactID := uuid.NewV7()
	now := time.Now()
	later := now.Add(time.Hour)
	text := "hello"
	existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusPending, []Message{*existingMsg}, nil, &contactID, now, nil, nil)

	newMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, now, nil, nil, nil)
	newMsg.AssignExternalID("wa-002")

	// Act
	err := conv.ReceiveContactMessage(contactID, *newMsg, later)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &later, conv.UpdatedAt())
}
