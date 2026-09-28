package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewConversation(t *testing.T) {
	// Arrange
	validID := uuid.NewV7()
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
	messages := []*Message{msg}

	tests := []struct {
		name      string
		id        uuid.UUID
		status    ConversationStatus
		messages  []*Message
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
			messages:  []*Message{},
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
		{
			name:      "expired without finishedAt",
			id:        validID,
			status:    ConversationStatusExpired,
			messages:  messages,
			contactID: &contactID,
			createdAt: now,
			wantErr:   ErrConversationFinishedAtMissing,
		},
		{
			name:      "resolved without finishedAt",
			id:        validID,
			status:    ConversationStatusResolved,
			messages:  messages,
			contactID: &contactID,
			createdAt: now,
			wantErr:   ErrConversationFinishedAtMissing,
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
	msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)

	// Act
	conv, err := NewConversation(id, ConversationStatusAssigned, []*Message{msg}, &agentID, &contactID, now, nil, nil)

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
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"
	msg1, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
	msg2, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)

	// Act
	conv, err := NewConversation(convID, ConversationStatusAssigned, []*Message{msg1, msg2}, &agentID, nil, now, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conv.Messages(), 2)
}

func TestConversation_ReceiveContactMessage(t *testing.T) {
	// Arrange
	contactID := uuid.NewV7()
	otherContact := uuid.NewV7()
	agentID := uuid.NewV7()
	now := time.Now()
	later := now.Add(time.Hour)
	finishedAt := now.Add(-time.Hour)
	lateFinishedAt := later
	text := "hello"

	tests := []struct {
		name        string
		convStatus  ConversationStatus
		agentID     *uuid.UUID
		contactID   *uuid.UUID
		finishedAt  *time.Time
		newMsgFunc  func() *Message
		receiveFrom uuid.UUID
		receiveAt   time.Time
		wantErr     error
		wantLen     int
	}{
		{
			name:       "success",
			convStatus: ConversationStatusPending,
			contactID:  &contactID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				_ = msg.AssignExternalID("wa-002")
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   now,
			wantErr:     nil,
			wantLen:     2,
		},
		{
			name:       "contact not owner",
			convStatus: ConversationStatusPending,
			contactID:  &contactID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &otherContact, nil, now, nil, nil, nil)
				return msg
			},
			receiveFrom: otherContact,
			receiveAt:   now,
			wantErr:     ErrConversationContactNotOwner,
			wantLen:     1,
		},
		{
			name:       "conversation has no contact",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			contactID:  nil,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   now,
			wantErr:     ErrConversationHasNoContact,
			wantLen:     1,
		},
		{
			name:       "duplicate message",
			convStatus: ConversationStatusPending,
			contactID:  &contactID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				_ = msg.AssignExternalID("wa-001")
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   now,
			wantErr:     ErrConversationDuplicateMessage,
			wantLen:     1,
		},
		{
			name:       "without external ID",
			convStatus: ConversationStatusPending,
			contactID:  &contactID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   now,
			wantErr:     nil,
			wantLen:     2,
		},
		{
			name:       "updates timestamp",
			convStatus: ConversationStatusPending,
			contactID:  &contactID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				_ = msg.AssignExternalID("wa-002")
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   later,
			wantErr:     nil,
			wantLen:     2,
		},
		{
			name:       "expired conversation rejects late message",
			convStatus: ConversationStatusExpired,
			contactID:  &contactID,
			finishedAt: &finishedAt,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   now,
			wantErr:     ErrConversationNotAcceptingMessages,
			wantLen:     1,
		},
		{
			name:       "resolved conversation rejects late message",
			convStatus: ConversationStatusResolved,
			contactID:  &contactID,
			finishedAt: &finishedAt,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   now,
			wantErr:     ErrConversationNotAcceptingMessages,
			wantLen:     1,
		},
		{
			name:       "expired conversation accepts message before finishedAt",
			convStatus: ConversationStatusExpired,
			contactID:  &contactID,
			finishedAt: &lateFinishedAt,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return msg
			},
			receiveFrom: contactID,
			receiveAt:   now,
			wantErr:     nil,
			wantLen:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			_ = existingMsg.AssignExternalID("wa-001")
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{existingMsg}, tt.agentID, tt.contactID, now, nil, tt.finishedAt)

			newMsg := tt.newMsgFunc()

			// Act
			err := conv.ReceiveContactMessage(tt.receiveFrom, newMsg, tt.receiveAt)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				if tt.name == "updates timestamp" {
					assert.Equal(t, &tt.receiveAt, conv.UpdatedAt())
				}
			}
			assert.Len(t, conv.Messages(), tt.wantLen)
		})
	}
}

func TestConversation_SendAgentMessage(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	otherAgent := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	finishedAt := now.Add(-time.Hour)
	text := "hello"

	tests := []struct {
		name       string
		convStatus ConversationStatus
		agentID    *uuid.UUID
		finishedAt *time.Time
		newMsgFunc func() *Message
		sendAgent  uuid.UUID
		wantErr    error
		wantLen    int
		wantAgent  *uuid.UUID
		wantStatus ConversationStatus
	}{
		{
			name:       "success",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			sendAgent:  agentID,
			wantErr:    nil,
			wantLen:    2,
			wantAgent:  &agentID,
			wantStatus: ConversationStatusAssigned,
		},
		{
			name:       "claim pending conversation",
			convStatus: ConversationStatusPending,
			agentID:    nil,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			sendAgent:  agentID,
			wantErr:    nil,
			wantLen:    2,
			wantAgent:  &agentID,
			wantStatus: ConversationStatusAssigned,
		},
		{
			name:       "agent not owner",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &otherAgent, nil, nil, now, nil, nil, nil)
				return msg
			},
			sendAgent:  otherAgent,
			wantErr:    ErrConversationAgentNotOwner,
			wantLen:    1,
			wantAgent:  &agentID,
			wantStatus: ConversationStatusAssigned,
		},
		{
			name:       "sync external ID",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
				_ = msg.AssignExternalID("wa-agent-001")
				return msg
			},
			sendAgent:  agentID,
			wantErr:    nil,
			wantLen:    2,
			wantAgent:  &agentID,
			wantStatus: ConversationStatusAssigned,
		},
		{
			name:       "without external ID",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			sendAgent:  agentID,
			wantErr:    nil,
			wantLen:    2,
			wantAgent:  &agentID,
			wantStatus: ConversationStatusAssigned,
		},
		{
			name:       "expired conversation rejects late message",
			convStatus: ConversationStatusExpired,
			agentID:    &agentID,
			finishedAt: &finishedAt,
			newMsgFunc: func() *Message {
				msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
				return msg
			},
			sendAgent:  agentID,
			wantErr:    ErrConversationNotAcceptingMessages,
			wantLen:    1,
			wantAgent:  &agentID,
			wantStatus: ConversationStatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{existingMsg}, tt.agentID, &contactID, now, nil, tt.finishedAt)

			newMsg := tt.newMsgFunc()

			// Act
			err := conv.SendAgentMessage(tt.sendAgent, newMsg, now)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &now, conv.UpdatedAt())
			}
			assert.Len(t, conv.Messages(), tt.wantLen)
			assert.Equal(t, tt.wantAgent, conv.AgentID())
			assert.Equal(t, tt.wantStatus, conv.Status())
		})
	}
}

func TestConversation_AgentReadConversation(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	otherAgent := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	later := now.Add(time.Hour)
	finishedAt := now.Add(-time.Hour)
	text := "hello"

	tests := []struct {
		name          string
		convStatus    ConversationStatus
		convAgentID   *uuid.UUID
		finishedAt    *time.Time
		readAgent     uuid.UUID
		readAt        time.Time
		setupMsgs     func() []*Message
		wantErr       error
		wantReadCount int
	}{
		{
			name:        "success - mark contact messages as read",
			convStatus:  ConversationStatusAssigned,
			convAgentID: &agentID,
			readAgent:   agentID,
			readAt:      now,
			setupMsgs: func() []*Message {
				contactMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return []*Message{contactMsg}
			},
			wantErr:       nil,
			wantReadCount: 1,
		},
		{
			name:        "fail - agent not owner",
			convStatus:  ConversationStatusAssigned,
			convAgentID: &agentID,
			readAgent:   otherAgent,
			readAt:      now,
			setupMsgs: func() []*Message {
				contactMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return []*Message{contactMsg}
			},
			wantErr:       ErrConversationAgentNotOwner,
			wantReadCount: 0,
		},
		{
			name:        "do not mark agent messages as read",
			convStatus:  ConversationStatusAssigned,
			convAgentID: &agentID,
			readAgent:   agentID,
			readAt:      now,
			setupMsgs: func() []*Message {
				agentMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
				contactMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return []*Message{agentMsg, contactMsg}
			},
			wantErr:       nil,
			wantReadCount: 1,
		},
		{
			name:        "mark messages in finished conversation",
			convStatus:  ConversationStatusExpired,
			convAgentID: &agentID,
			finishedAt:  &finishedAt,
			readAgent:   agentID,
			readAt:      now,
			setupMsgs: func() []*Message {
				contactMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return []*Message{contactMsg}
			},
			wantErr:       nil,
			wantReadCount: 1,
		},
		{
			name:        "do not mark already read messages",
			convStatus:  ConversationStatusAssigned,
			convAgentID: &agentID,
			readAgent:   agentID,
			readAt:      later,
			setupMsgs: func() []*Message {
				readMsg, _ := NewMessage(uuid.NewV7(), MessageStatusRead, MessageTypeText, &text, nil, &contactID, nil, now, &now, nil, nil)
				unreadMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
				return []*Message{readMsg, unreadMsg}
			},
			wantErr:       nil,
			wantReadCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgs := tt.setupMsgs()
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, msgs, tt.convAgentID, &contactID, now, nil, tt.finishedAt)

			// Act
			err := conv.AgentReadConversation(tt.readAgent, tt.readAt)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.readAt, conv.UpdatedAt())
			}

			readCount := 0
			for _, msg := range conv.Messages() {
				if msg.ReadAt() != nil {
					readCount++
				}
			}
			assert.Equal(t, tt.wantReadCount, readCount)
		})
	}
}

func TestConversation_ReceiveContactMessage_DoesNotRetrocedeUpdatedAt(t *testing.T) {
	// Arrange
	contactID := uuid.NewV7()
	now := time.Now()
	later := now.Add(time.Hour)
	text := "hello"
	existingMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusPending, []*Message{existingMsg}, nil, &contactID, now, &later, nil)
	newMsg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)

	// Act
	err := conv.ReceiveContactMessage(contactID, newMsg, now)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &later, conv.UpdatedAt())
}

func TestConversation_MarkAgentMessageFailed(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	otherAgent := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	finishedAt := now.Add(-time.Hour)
	text := "hello"
	agentMsgID := uuid.NewV7()
	contactMsgID := uuid.NewV7()

	tests := []struct {
		name       string
		convStatus ConversationStatus
		finishedAt *time.Time
		msgID      uuid.UUID
		callerID   uuid.UUID
		wantErr    error
	}{
		{
			name:       "success",
			convStatus: ConversationStatusAssigned,
			msgID:      agentMsgID,
			callerID:   agentID,
			wantErr:    nil,
		},
		{
			name:       "agent not owner",
			convStatus: ConversationStatusAssigned,
			msgID:      agentMsgID,
			callerID:   otherAgent,
			wantErr:    ErrConversationAgentNotOwner,
		},
		{
			name:       "message not found",
			convStatus: ConversationStatusAssigned,
			msgID:      uuid.NewV7(),
			callerID:   agentID,
			wantErr:    ErrMessageNotFound,
		},
		{
			name:       "message not from agent",
			convStatus: ConversationStatusAssigned,
			msgID:      contactMsgID,
			callerID:   agentID,
			wantErr:    ErrMessageNotFromAgent,
		},
		{
			name:       "conversation finished is allowed",
			convStatus: ConversationStatusExpired,
			finishedAt: &finishedAt,
			msgID:      agentMsgID,
			callerID:   agentID,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentMsg, _ := NewMessage(agentMsgID, MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
			contactMsg, _ := NewMessage(contactMsgID, MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{agentMsg, contactMsg}, &agentID, &contactID, now, nil, tt.finishedAt)

			// Act
			err := conv.MarkAgentMessageFailed(tt.callerID, tt.msgID, now)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, MessageStatusFailed, conv.found[tt.msgID].Status())
		})
	}
}

func TestConversation_ReceiveContactMessageEdit(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	otherContact := uuid.NewV7()
	now := time.Now()
	finishedAt := now.Add(-time.Hour)
	text := "hello"
	newText := "edited"
	contactMsgID := uuid.NewV7()
	agentMsgID := uuid.NewV7()

	const (
		contactExternalID = "wa-contact-001"
		agentExternalID   = "wa-agent-001"
	)

	tests := []struct {
		name       string
		convStatus ConversationStatus
		finishedAt *time.Time
		externalID string
		callerID   uuid.UUID
		msgStatus  MessageStatus
		wantErr    error
	}{
		{
			name:       "success",
			convStatus: ConversationStatusAssigned,
			externalID: contactExternalID,
			callerID:   contactID,
			wantErr:    nil,
		},
		{
			name:       "contact not owner",
			convStatus: ConversationStatusAssigned,
			externalID: contactExternalID,
			callerID:   otherContact,
			wantErr:    ErrConversationContactNotOwner,
		},
		{
			name:       "externalID not found",
			convStatus: ConversationStatusAssigned,
			externalID: "wa-unknown",
			callerID:   contactID,
			wantErr:    ErrMessageNotFound,
		},
		{
			name:       "message not from contact",
			convStatus: ConversationStatusAssigned,
			externalID: agentExternalID,
			callerID:   contactID,
			wantErr:    ErrConversationContactNotOwner,
		},
		{
			name:       "message failed",
			convStatus: ConversationStatusAssigned,
			externalID: contactExternalID,
			callerID:   contactID,
			msgStatus:  MessageStatusFailed,
			wantErr:    ErrMessageFailed,
		},
		{
			name:       "conversation finished is allowed",
			convStatus: ConversationStatusExpired,
			finishedAt: &finishedAt,
			externalID: contactExternalID,
			callerID:   contactID,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.msgStatus
			if status == "" {
				status = MessageStatusSent
			}

			contactMsg, _ := NewMessage(contactMsgID, status, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			_ = contactMsg.AssignExternalID(contactExternalID)
			agentMsg, _ := NewMessage(agentMsgID, MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
			_ = agentMsg.AssignExternalID(agentExternalID)
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{contactMsg, agentMsg}, &agentID, &contactID, now, nil, tt.finishedAt)

			// Act
			err := conv.ReceiveContactMessageEdit(tt.callerID, tt.externalID, newText, now)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, newText, *conv.found[contactMsgID].Text())
			assert.Equal(t, &now, conv.found[contactMsgID].EditedAt())
		})
	}
}

func TestConversation_ReceiveContactMessageDelete(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	otherContact := uuid.NewV7()
	now := time.Now()
	finishedAt := now.Add(-time.Hour)
	text := "hello"
	contactMsgID := uuid.NewV7()
	agentMsgID := uuid.NewV7()

	const (
		contactExternalID = "wa-contact-001"
		agentExternalID   = "wa-agent-001"
	)

	tests := []struct {
		name       string
		convStatus ConversationStatus
		finishedAt *time.Time
		externalID string
		callerID   uuid.UUID
		msgStatus  MessageStatus
		wantErr    error
	}{
		{
			name:       "success",
			convStatus: ConversationStatusAssigned,
			externalID: contactExternalID,
			callerID:   contactID,
			wantErr:    nil,
		},
		{
			name:       "contact not owner",
			convStatus: ConversationStatusAssigned,
			externalID: contactExternalID,
			callerID:   otherContact,
			wantErr:    ErrConversationContactNotOwner,
		},
		{
			name:       "externalID not found",
			convStatus: ConversationStatusAssigned,
			externalID: "wa-unknown",
			callerID:   contactID,
			wantErr:    ErrMessageNotFound,
		},
		{
			name:       "message not from contact",
			convStatus: ConversationStatusAssigned,
			externalID: agentExternalID,
			callerID:   contactID,
			wantErr:    ErrConversationContactNotOwner,
		},
		{
			name:       "message failed",
			convStatus: ConversationStatusAssigned,
			externalID: contactExternalID,
			callerID:   contactID,
			msgStatus:  MessageStatusFailed,
			wantErr:    ErrMessageFailed,
		},
		{
			name:       "conversation finished is allowed",
			convStatus: ConversationStatusExpired,
			finishedAt: &finishedAt,
			externalID: contactExternalID,
			callerID:   contactID,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.msgStatus
			if status == "" {
				status = MessageStatusSent
			}

			contactMsg, _ := NewMessage(contactMsgID, status, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			_ = contactMsg.AssignExternalID(contactExternalID)
			agentMsg, _ := NewMessage(agentMsgID, MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)
			_ = agentMsg.AssignExternalID(agentExternalID)
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{contactMsg, agentMsg}, &agentID, &contactID, now, nil, tt.finishedAt)

			// Act
			err := conv.ReceiveContactMessageDelete(tt.callerID, tt.externalID, now)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, MessageStatusDeleted, conv.found[contactMsgID].Status())
			assert.Equal(t, &now, conv.found[contactMsgID].DeletedAt())
		})
	}
}

func TestConversation_ReceiveContactMessageRead(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	otherContact := uuid.NewV7()
	now := time.Now()
	newestReadAt := now.Add(time.Hour)
	finishedAt := now.Add(-time.Hour)
	deletedAt := now.Add(-time.Hour)
	text := "hello"
	contactMsgID := uuid.NewV7()
	agentMsgID := uuid.NewV7()

	const (
		contactExternalID = "wa-contact-001"
		agentExternalID   = "wa-agent-001"
	)

	tests := []struct {
		name        string
		convStatus  ConversationStatus
		finishedAt  *time.Time
		externalID  string
		callerID    uuid.UUID
		agentStatus MessageStatus
		preReadAt   *time.Time
		deletedAt   *time.Time
		wantErr     error
		wantStatus  MessageStatus
		wantReadAt  *time.Time
	}{
		{
			name:       "success",
			convStatus: ConversationStatusAssigned,
			externalID: agentExternalID,
			callerID:   contactID,
			wantStatus: MessageStatusRead,
			wantReadAt: &now,
		},
		{
			name:       "contact not owner",
			convStatus: ConversationStatusAssigned,
			externalID: agentExternalID,
			callerID:   otherContact,
			wantErr:    ErrConversationContactNotOwner,
		},
		{
			name:       "externalID not found",
			convStatus: ConversationStatusAssigned,
			externalID: "wa-unknown",
			callerID:   contactID,
			wantErr:    ErrMessageNotFound,
		},
		{
			name:       "message not from agent",
			convStatus: ConversationStatusAssigned,
			externalID: contactExternalID,
			callerID:   contactID,
			wantErr:    ErrConversationContactNotOwner,
		},
		{
			name:       "conversation finished is allowed",
			convStatus: ConversationStatusExpired,
			finishedAt: &finishedAt,
			externalID: agentExternalID,
			callerID:   contactID,
			wantStatus: MessageStatusRead,
			wantReadAt: &now,
		},
		{
			name:        "stale read is discarded",
			convStatus:  ConversationStatusAssigned,
			externalID:  agentExternalID,
			callerID:    contactID,
			agentStatus: MessageStatusRead,
			preReadAt:   &newestReadAt,
			wantStatus:  MessageStatusRead,
			wantReadAt:  &newestReadAt,
		},
		{
			name:        "deleted message keeps deleted status and records read",
			convStatus:  ConversationStatusAssigned,
			externalID:  agentExternalID,
			callerID:    contactID,
			agentStatus: MessageStatusDeleted,
			deletedAt:   &deletedAt,
			wantStatus:  MessageStatusDeleted,
			wantReadAt:  &now,
		},
		{
			name:        "failed message is not marked as read",
			convStatus:  ConversationStatusAssigned,
			externalID:  agentExternalID,
			callerID:    contactID,
			agentStatus: MessageStatusFailed,
			wantStatus:  MessageStatusFailed,
			wantReadAt:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contactMsg, _ := NewMessage(contactMsgID, MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			_ = contactMsg.AssignExternalID(contactExternalID)

			agentStatus := tt.agentStatus
			if agentStatus == "" {
				agentStatus = MessageStatusSent
			}
			agentMsg, _ := NewMessage(agentMsgID, agentStatus, MessageTypeText, &text, &agentID, nil, nil, now, tt.preReadAt, nil, tt.deletedAt)
			_ = agentMsg.AssignExternalID(agentExternalID)

			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{contactMsg, agentMsg}, &agentID, &contactID, now, nil, tt.finishedAt)

			// Act
			err := conv.ReceiveContactMessageRead(tt.callerID, tt.externalID, now)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, conv.found[agentMsgID].Status())
			assert.Equal(t, tt.wantReadAt, conv.found[agentMsgID].ReadAt())
		})
	}
}

func TestConversation_ReceiveContactMessageEdit_DiscardsStaleEdit(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	newestAt := now.Add(time.Hour)
	text := "hello"
	externalID := "wa-contact-001"
	msgID := uuid.NewV7()

	contactMsg, _ := NewMessage(msgID, MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
	_ = contactMsg.AssignExternalID(externalID)
	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusAssigned, []*Message{contactMsg}, &agentID, &contactID, now, nil, nil)
	_ = conv.ReceiveContactMessageEdit(contactID, externalID, "newest", newestAt)

	// Act
	err := conv.ReceiveContactMessageEdit(contactID, externalID, "stale", now)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "newest", *conv.found[msgID].Text())
	assert.Equal(t, &newestAt, conv.found[msgID].EditedAt())
}

func TestConversation_ExpireConversation(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	at := now.Add(time.Hour)
	earlier := now.Add(-time.Hour)
	text := "hello"

	tests := []struct {
		name       string
		convStatus ConversationStatus
		agentID    *uuid.UUID
		updatedAt  *time.Time
		finishedAt *time.Time
		wantErr    error
		wantStatus ConversationStatus
	}{
		{
			name:       "expire pending",
			convStatus: ConversationStatusPending,
			wantStatus: ConversationStatusExpired,
		},
		{
			name:       "expire assigned",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			wantStatus: ConversationStatusExpired,
		},
		{
			name:       "fail already expired",
			convStatus: ConversationStatusExpired,
			agentID:    &agentID,
			finishedAt: &earlier,
			wantErr:    ErrConversationFinished,
		},
		{
			name:       "fail already resolved",
			convStatus: ConversationStatusResolved,
			agentID:    &agentID,
			finishedAt: &earlier,
			wantErr:    ErrConversationFinished,
		},
		{
			name:       "registers activity",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			updatedAt:  &earlier,
			wantStatus: ConversationStatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{msg}, tt.agentID, &contactID, now, tt.updatedAt, tt.finishedAt)

			// Act
			err := conv.ExpireConversation(at)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, conv.Status())
			assert.Equal(t, &at, conv.FinishedAt())
			assert.Equal(t, &at, conv.UpdatedAt())
		})
	}
}

func TestConversation_ResolveConversation(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	otherAgent := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	at := now.Add(time.Hour)
	earlier := now.Add(-time.Hour)
	text := "hello"

	tests := []struct {
		name       string
		convStatus ConversationStatus
		agentID    *uuid.UUID
		updatedAt  *time.Time
		finishedAt *time.Time
		callerID   uuid.UUID
		wantErr    error
		wantStatus ConversationStatus
	}{
		{
			name:       "resolve assigned",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			callerID:   agentID,
			wantStatus: ConversationStatusResolved,
		},
		{
			name:       "fail agent not owner",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			callerID:   otherAgent,
			wantErr:    ErrConversationAgentNotOwner,
		},
		{
			name:       "fail conversation has no agent",
			convStatus: ConversationStatusPending,
			callerID:   agentID,
			wantErr:    ErrConversationAgentNotOwner,
		},
		{
			name:       "fail already resolved",
			convStatus: ConversationStatusResolved,
			agentID:    &agentID,
			finishedAt: &earlier,
			callerID:   agentID,
			wantErr:    ErrConversationFinished,
		},
		{
			name:       "fail already expired",
			convStatus: ConversationStatusExpired,
			agentID:    &agentID,
			finishedAt: &earlier,
			callerID:   agentID,
			wantErr:    ErrConversationFinished,
		},
		{
			name:       "registers activity",
			convStatus: ConversationStatusAssigned,
			agentID:    &agentID,
			updatedAt:  &earlier,
			callerID:   agentID,
			wantStatus: ConversationStatusResolved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
			conv, _ := NewConversation(uuid.NewV7(), tt.convStatus, []*Message{msg}, tt.agentID, &contactID, now, tt.updatedAt, tt.finishedAt)

			// Act
			err := conv.ResolveConversation(tt.callerID, at)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, conv.Status())
			assert.Equal(t, &at, conv.FinishedAt())
			assert.Equal(t, &at, conv.UpdatedAt())
		})
	}
}

func TestConversation_Messages_OrderedBySentAt(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"

	msgLate, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now.Add(2*time.Hour), nil, nil, nil)
	msgEarly, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
	msgMid, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now.Add(time.Hour), nil, nil, nil)

	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusAssigned, []*Message{msgLate, msgEarly, msgMid}, &agentID, &contactID, now, nil, nil)

	// Act
	msgs := conv.Messages()

	// Assert
	assert.Len(t, msgs, 3)
	assert.Equal(t, msgEarly.ID(), msgs[0].ID())
	assert.Equal(t, msgMid.ID(), msgs[1].ID())
	assert.Equal(t, msgLate.ID(), msgs[2].ID())
}

func TestConversation_Messages_TieBreakByID(t *testing.T) {
	// Arrange
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	text := "hello"

	msgA, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, now, nil, nil, nil)
	msgB, _ := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, now, nil, nil, nil)

	conv, _ := NewConversation(uuid.NewV7(), ConversationStatusAssigned, []*Message{msgB, msgA}, &agentID, &contactID, now, nil, nil)

	first, second := msgA, msgB
	if msgB.ID().Compare(msgA.ID()) < 0 {
		first, second = msgB, msgA
	}

	// Act
	msgs := conv.Messages()

	// Assert
	assert.Len(t, msgs, 2)
	assert.Equal(t, first.ID(), msgs[0].ID())
	assert.Equal(t, second.ID(), msgs[1].ID())
}

func TestRehydrateConversation(t *testing.T) {
	convID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	externalID := "wa-contact-001"

	tests := []struct {
		name       string
		id         uuid.UUID
		status     ConversationStatus
		messages   []*Message
		agentID    *uuid.UUID
		contactID  *uuid.UUID
		finishedAt *time.Time
		wantLen    int
	}{
		{
			name:      "hydrates without messages",
			id:        convID,
			status:    ConversationStatusPending,
			messages:  nil,
			contactID: &contactID,
			wantLen:   0,
		},
		{
			name:      "hydrates with messages",
			id:        convID,
			status:    ConversationStatusAssigned,
			messages:  []*Message{mustMessage(t, contactID)},
			contactID: &contactID,
			wantLen:   1,
		},
		{
			name:      "maps an invalid state directly",
			id:        uuid.Nil(),
			status:    ConversationStatusResolved,
			contactID: nil,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			conv := RehydrateConversation(tt.id, tt.status, tt.messages, tt.agentID, tt.contactID, now, nil, tt.finishedAt)

			// Assert
			assert.NotNil(t, conv)
			assert.Equal(t, tt.id, conv.ID())
			assert.Equal(t, tt.status, conv.Status())
			assert.Len(t, conv.Messages(), tt.wantLen)
		})
	}

	t.Run("builds the external message index", func(t *testing.T) {
		// Arrange
		message := mustMessageWithExternalID(t, contactID, externalID)

		// Act
		conv := RehydrateConversation(convID, ConversationStatusPending, []*Message{message}, nil, &contactID, now, nil, nil)

		// Assert
		assert.Equal(t, message.ID(), conv.externalMsgIdx[externalID])
	})
}

func mustMessageWithExternalID(t *testing.T, contactID uuid.UUID, externalID string) *Message {
	t.Helper()
	text := "hello"
	msg, err := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, &externalID, time.Now(), nil, nil, nil)
	assert.NoError(t, err)
	return msg
}

func TestConversation_AssignAgentMessageExternalID(t *testing.T) {
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name        string
		msgID       func(agentMsg *Message, contactMsg *Message) uuid.UUID
		externalID  string
		preassigned string
		wantErr     error
	}{
		{
			name:       "assign external id to agent message",
			msgID:      func(agentMsg *Message, contactMsg *Message) uuid.UUID { return agentMsg.ID() },
			externalID: "wa-agent-001",
			wantErr:    nil,
		},
		{
			name:       "message not found",
			msgID:      func(agentMsg *Message, contactMsg *Message) uuid.UUID { return uuid.NewV7() },
			externalID: "wa-agent-002",
			wantErr:    ErrMessageNotFound,
		},
		{
			name:       "not an agent message",
			msgID:      func(agentMsg *Message, contactMsg *Message) uuid.UUID { return contactMsg.ID() },
			externalID: "wa-agent-003",
			wantErr:    ErrConversationAgentNotOwner,
		},
		{
			name:        "external id already assigned",
			msgID:       func(agentMsg *Message, contactMsg *Message) uuid.UUID { return agentMsg.ID() },
			externalID:  "wa-agent-004",
			preassigned: "wa-agent-existing",
			wantErr:     ErrMessageExternalIDAlreadySet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			agentMsg := mustMessageWithAgent(t, agentID)
			if tt.preassigned != "" {
				assert.NoError(t, agentMsg.AssignExternalID(tt.preassigned))
			}
			contactMsg := mustMessage(t, contactID)
			conv, err := NewConversation(uuid.NewV7(), ConversationStatusAssigned, []*Message{agentMsg, contactMsg}, &agentID, &contactID, now, nil, nil)
			assert.NoError(t, err)

			// Act
			err = conv.AssignAgentMessageExternalID(tt.msgID(agentMsg, contactMsg), tt.externalID, now)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, agentMsg.ExternalID())
			assert.Equal(t, tt.externalID, *agentMsg.ExternalID())
			assert.Equal(t, agentMsg.ID(), conv.externalMsgIdx[tt.externalID])
		})
	}
}

func mustMessage(t *testing.T, contactID uuid.UUID) *Message {
	t.Helper()
	text := "hello"
	msg, err := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, nil, &contactID, nil, time.Now(), nil, nil, nil)
	assert.NoError(t, err)
	return msg
}

func mustMessageWithAgent(t *testing.T, agentID uuid.UUID) *Message {
	t.Helper()
	text := "hello"
	msg, err := NewMessage(uuid.NewV7(), MessageStatusSent, MessageTypeText, &text, &agentID, nil, nil, time.Now(), nil, nil, nil)
	assert.NoError(t, err)
	return msg
}
