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
			agentID:   &agentID,
			createdAt: now,
			wantErr:   nil,
		},
		{
			name:      "valid with contact",
			id:        validID,
			status:    ConversationStatusPending,
			contactID: &contactID,
			createdAt: now,
			wantErr:   nil,
		},
		{
			name:      "valid with both",
			id:        validID,
			status:    ConversationStatusAssigned,
			agentID:   &agentID,
			contactID: &contactID,
			createdAt: now,
			wantErr:   nil,
		},
		{
			name:      "nil uuid",
			id:        uuid.Nil(),
			status:    ConversationStatusPending,
			contactID: &contactID,
			createdAt: now,
			wantErr:   ErrConversationInvalidID,
		},
		{
			name:      "missing contact and agent",
			id:        validID,
			status:    ConversationStatusPending,
			createdAt: now,
			wantErr:   ErrConversationMissingContactAndAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv, err := NewConversation(tt.id, tt.status, tt.messages, tt.agentID, tt.contactID, tt.createdAt, nil, nil)
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
	id := uuid.NewV7()
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()

	conv, err := NewConversation(id, ConversationStatusAssigned, nil, &agentID, &contactID, now, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, id, conv.ID())
	assert.Equal(t, ConversationStatusAssigned, conv.Status())
	assert.Equal(t, &agentID, conv.AgentID())
	assert.Equal(t, &contactID, conv.ContactID())
	assert.Equal(t, now, conv.CreatedAt())
	assert.Nil(t, conv.UpdatedAt())
	assert.Nil(t, conv.FinishedAt())
	assert.Empty(t, conv.Messages())
}

func TestConversation_Messages(t *testing.T) {
	convID := uuid.NewV7()
	agentID := uuid.NewV7()
	now := time.Now()
	text := "hello"

	msg1, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, nil, now, nil, nil, nil)
	msg2, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, nil, now, nil, nil, nil)

	messages := []Message{*msg1, *msg2}

	conv, err := NewConversation(convID, ConversationStatusAssigned, messages, &agentID, nil, now, nil, nil)
	assert.NoError(t, err)

	result := conv.Messages()
	assert.Len(t, result, 2)
}
