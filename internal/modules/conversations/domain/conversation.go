package domain

import (
	"errors"
	"time"

	"uuid"
)

var (
	ErrConversationInvalidID              = errors.New("conversation identifier is invalid")
	ErrConversationMissingContactAndAgent = errors.New("conversation must have a contact or an agent")
)

type Conversation struct {
	id         uuid.UUID
	status     ConversationStatus
	found      map[uuid.UUID]Message
	dirty      map[uuid.UUID]Message
	agentID    *uuid.UUID
	contactID  *uuid.UUID
	createdAt  time.Time
	updatedAt  *time.Time
	finishedAt *time.Time
}

func NewConversation(
	id uuid.UUID,
	status ConversationStatus,
	messages []Message,
	agentID *uuid.UUID,
	contactID *uuid.UUID,
	createdAt time.Time,
	updatedAt *time.Time,
	finishedAt *time.Time,
) (*Conversation, error) {
	if id == uuid.Nil() {
		return nil, ErrConversationInvalidID
	}
	if agentID == nil && contactID == nil {
		return nil, ErrConversationMissingContactAndAgent
	}

	found := make(map[uuid.UUID]Message, len(messages))
	for _, msg := range messages {
		found[msg.ID()] = msg
	}

	return &Conversation{
		id:         id,
		status:     status,
		found:      found,
		dirty:      make(map[uuid.UUID]Message),
		agentID:    agentID,
		contactID:  contactID,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
		finishedAt: finishedAt,
	}, nil
}

func (c *Conversation) ID() uuid.UUID              { return c.id }
func (c *Conversation) Status() ConversationStatus { return c.status }
func (c *Conversation) AgentID() *uuid.UUID        { return c.agentID }
func (c *Conversation) ContactID() *uuid.UUID      { return c.contactID }
func (c *Conversation) CreatedAt() time.Time       { return c.createdAt }
func (c *Conversation) UpdatedAt() *time.Time      { return c.updatedAt }
func (c *Conversation) FinishedAt() *time.Time     { return c.finishedAt }
func (c *Conversation) Messages() []Message {
	msgs := make([]Message, 0, len(c.found))
	for _, msg := range c.found {
		msgs = append(msgs, msg)
	}
	return msgs
}
