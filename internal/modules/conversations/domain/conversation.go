package domain

import (
	"errors"
	"time"
	"uuid"
)

const (
	maxUnreadCount int8 = 100
	minUnreadCount int8 = 0
)

var (
	ErrUnauthorizedAgent       = errors.New("unauthorized agent")
	ErrUnauthorizedContact     = errors.New("unauthorized contact")
	ErrMessageNotFound         = errors.New("message not found")
	ErrConversationClosed      = errors.New("conversation closed")
	ErrConversationNotFound    = errors.New("conversation not found")
	ErrMessageAlreadyExists    = errors.New("message already exists")
	ErrConversationNotAssigned = errors.New("conversation not assigned")
)

type Conversation struct {
	id              uuid.UUID
	status          ConversationStatus
	foundMessages   map[uuid.UUID]*Message
	addedMessages   map[uuid.UUID]*Message
	updatedMessages map[uuid.UUID]*Message
	deletedMessages map[uuid.UUID]*Message
	unreadCount     int8
	agentID         *uuid.UUID
	contactID       uuid.UUID
	createdAt       time.Time
	updatedAt       *time.Time
	finishedAt      *time.Time
}

func NewConversation(
	id uuid.UUID,
	status ConversationStatus,
	foundMessages map[uuid.UUID]*Message,
	unreadCount int8,
	agentID *uuid.UUID,
	contactID uuid.UUID,
	createdAt time.Time,
	updatedAt *time.Time,
	finishedAt *time.Time,
) (*Conversation, error) {
	if foundMessages == nil {
		foundMessages = make(map[uuid.UUID]*Message)
	}

	return &Conversation{
		id,
		status,
		foundMessages,
		make(map[uuid.UUID]*Message),
		make(map[uuid.UUID]*Message),
		make(map[uuid.UUID]*Message),
		unreadCount,
		agentID,
		contactID,
		createdAt,
		updatedAt,
		finishedAt,
	}, nil
}

func (c *Conversation) ID() uuid.UUID              { return c.id }
func (c *Conversation) Status() ConversationStatus { return c.status }
func (c *Conversation) UnreadCount() int8          { return c.unreadCount }
func (c *Conversation) AgentID() *uuid.UUID        { return c.agentID }
func (c *Conversation) ContactID() uuid.UUID       { return c.contactID }
func (c *Conversation) CreatedAt() time.Time       { return c.createdAt }
func (c *Conversation) UpdatedAt() *time.Time      { return c.updatedAt }
func (c *Conversation) FinishedAt() *time.Time     { return c.finishedAt }
func (c *Conversation) hasExternalID(externalID string) bool {
	for _, message := range c.foundMessages {
		extID := message.ExternalID()
		if extID != nil && *extID == externalID {
			return true
		}
	}
	return false
}
func (c *Conversation) findContactMessageByExternalID(contactID uuid.UUID, externalMessageID string) *Message {
	var found *Message
	for _, message := range c.foundMessages {
		if message.ExternalID() != nil && *message.ExternalID() == externalMessageID &&
			message.ContactID() != nil && *message.ContactID() == contactID {
			found = message
			break
		}
	}

	return found
}
func (c *Conversation) IsClosed(at time.Time) bool {
	return (c.status == Expired || c.status == Resolved) &&
		c.finishedAt != nil && c.finishedAt.Before(at)
}
func (c *Conversation) AgentSendMessage(messageID, agentID uuid.UUID, text string, at time.Time) error {
	if c.IsClosed(at) {
		return ErrConversationClosed
	}

	if c.agentID == nil {
		c.agentID = &agentID
		c.status = Assigned
		c.AgentReadConversation(agentID, at)
	}

	if *c.agentID != agentID {
		return ErrUnauthorizedAgent
	}

	message, err := NewMessage(messageID, nil, text, Registered, &agentID, nil, at, nil, nil, nil)
	if err != nil {
		return err
	}

	c.foundMessages[messageID] = message
	c.addedMessages[messageID] = message

	c.updatedAt = &at

	return nil
}
func (c *Conversation) AgentReadConversation(agentID uuid.UUID, at time.Time) error {
	if c.agentID == nil {
		return ErrConversationNotAssigned
	}

	if *c.agentID != agentID {
		return ErrUnauthorizedAgent
	}

	hasUpdates := false
	for _, message := range c.foundMessages {
		if message.ContactID() != nil && message.Read(at) {
			c.updatedMessages[message.ID()] = message
			hasUpdates = true
		}
	}

	if hasUpdates {
		c.unreadCount = minUnreadCount
		c.updatedAt = &at
	}

	return nil
}
func (c *Conversation) ReceiveContactMessage(messageID, contactID uuid.UUID, externalMessageID string, text string, at time.Time) error {
	if c.hasExternalID(externalMessageID) {
		return ErrMessageAlreadyExists
	}

	if c.contactID != contactID {
		return ErrUnauthorizedContact
	}

	if c.IsClosed(at) {
		return ErrConversationClosed
	}

	message, err := NewMessage(messageID, &externalMessageID, text, Delivered, nil, &contactID, at, nil, nil, nil)
	if err != nil {
		return err
	}

	c.foundMessages[messageID] = message
	c.addedMessages[messageID] = message

	c.addUnread(at)

	return nil
}
func (c *Conversation) ContactUpdateTextMessage(contactID uuid.UUID, externalMessageID, text string, at time.Time) error {
	if c.IsClosed(at) {
		return ErrConversationClosed
	}

	if c.contactID != contactID {
		return ErrUnauthorizedContact
	}

	message := c.findContactMessageByExternalID(contactID, externalMessageID)
	if message == nil {
		return ErrMessageNotFound
	}

	message.UpdateText(text, at)

	c.updatedMessages[message.ID()] = message
	c.updatedAt = &at

	return nil
}
func (c *Conversation) ContactDeleteMessage(contactID uuid.UUID, externalMessageID string, at time.Time) error {
	if c.IsClosed(at) {
		return ErrConversationClosed
	}

	if c.contactID != contactID {
		return ErrUnauthorizedContact
	}

	message := c.findContactMessageByExternalID(contactID, externalMessageID)
	if message == nil {
		return ErrMessageNotFound
	}

	wasUnread := message.Status() != Read
	if message.Delete(at) {
		c.deletedMessages[message.ID()] = message

		if wasUnread {
			c.substractUnread(at)
		}

		c.updatedAt = &at
	}

	return nil
}
func (c *Conversation) addUnread(at time.Time) {
	c.updatedAt = &at

	if c.unreadCount >= maxUnreadCount {
		c.unreadCount = maxUnreadCount
		return
	}

	c.unreadCount += 1
}
func (c *Conversation) substractUnread(at time.Time) {
	c.updatedAt = &at

	if c.unreadCount <= minUnreadCount {
		c.unreadCount = minUnreadCount
		return
	}

	c.unreadCount -= 1
}
