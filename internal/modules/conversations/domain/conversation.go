package domain

import (
	"errors"
	"slices"
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
	id            uuid.UUID
	status        ConversationStatus
	messages      map[uuid.UUID]*Message
	dirtyMessages map[uuid.UUID]*Message
	unreadCount   int8
	agentID       *uuid.UUID
	contactID     uuid.UUID
	createdAt     time.Time
	updatedAt     *time.Time
	finishedAt    *time.Time
}

func NewConversation(
	id uuid.UUID,
	status ConversationStatus,
	messages map[uuid.UUID]*Message,
	unreadCount int8,
	agentID *uuid.UUID,
	contactID uuid.UUID,
	createdAt time.Time,
	updatedAt *time.Time,
	finishedAt *time.Time,
) (*Conversation, error) {
	if messages == nil {
		messages = make(map[uuid.UUID]*Message)
	}

	return &Conversation{
		id,
		status,
		messages,
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
func (c *Conversation) Messages() []*Message       { return sortedByCreatedAt(c.messages) }
func (c *Conversation) DirtyMessages() []*Message  { return sortedByCreatedAt(c.dirtyMessages) }
func sortedByCreatedAt(messages map[uuid.UUID]*Message) []*Message {
	sorted := make([]*Message, 0, len(messages))
	for _, message := range messages {
		sorted = append(sorted, message)
	}

	slices.SortFunc(sorted, func(a, b *Message) int {
		return a.createdAt.Compare(b.createdAt)
	})

	return sorted
}
func (c *Conversation) updateUpdatedAt(at time.Time) {
	if c.updatedAt == nil || at.After(*c.updatedAt) {
		c.updatedAt = &at
	}
}
func (c *Conversation) hasExternalID(externalID string) bool {
	for _, message := range c.messages {
		extID := message.ExternalID()
		if extID != nil && *extID == externalID {
			return true
		}
	}
	return false
}
func (c *Conversation) findContactMessageByExternalID(contactID uuid.UUID, externalMessageID string) *Message {
	var found *Message
	for _, message := range c.messages {
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
func (c *Conversation) AgentSendMessage(messageID, agentID uuid.UUID, text string, at time.Time) (*Message, error) {
	if c.IsClosed(at) {
		return nil, ErrConversationClosed
	}

	if c.agentID == nil {
		c.agentID = &agentID
		c.status = Assigned
		c.AgentReadConversation(agentID, at)
	}

	if *c.agentID != agentID {
		return nil, ErrUnauthorizedAgent
	}

	message, err := NewMessage(messageID, nil, text, Registered, &agentID, nil, at, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	c.messages[messageID] = message
	c.dirtyMessages[messageID] = message

	c.updateUpdatedAt(at)

	return message, nil
}
func (c *Conversation) AgentReadConversation(agentID uuid.UUID, at time.Time) error {
	if c.agentID == nil {
		return ErrConversationNotAssigned
	}

	if *c.agentID != agentID {
		return ErrUnauthorizedAgent
	}

	hasUpdates := false
	for _, message := range c.messages {
		if message.ContactID() != nil && message.Read(at) {
			c.dirtyMessages[message.ID()] = message
			hasUpdates = true
		}
	}

	if hasUpdates {
		c.unreadCount = minUnreadCount
		c.updateUpdatedAt(at)
	}

	return nil
}
func (c *Conversation) AssignAgentMessageExternalID(messageID uuid.UUID, externalMessageID string) error {
	message, ok := c.messages[messageID]
	if !ok {
		return ErrMessageNotFound
	}

	if message.AssignExternalID(externalMessageID) {
		c.dirtyMessages[messageID] = message
	}

	return nil
}
func (c *Conversation) ReceiveContactMessage(messageID, contactID uuid.UUID, externalMessageID string, text string, at time.Time) (*Message, error) {
	if c.hasExternalID(externalMessageID) {
		return nil, ErrMessageAlreadyExists
	}

	if c.contactID != contactID {
		return nil, ErrUnauthorizedContact
	}

	if c.IsClosed(at) {
		return nil, ErrConversationClosed
	}

	message, err := NewMessage(messageID, &externalMessageID, text, Delivered, nil, &contactID, at, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	c.messages[messageID] = message
	c.dirtyMessages[messageID] = message

	c.addUnread(at)

	return message, nil
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

	c.dirtyMessages[message.ID()] = message
	c.updateUpdatedAt(at)

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
		c.dirtyMessages[message.ID()] = message

		if wasUnread {
			c.substractUnread(at)
		}

		c.updateUpdatedAt(at)
	}

	return nil
}
func (c *Conversation) addUnread(at time.Time) {
	c.updateUpdatedAt(at)

	if c.unreadCount >= maxUnreadCount {
		c.unreadCount = maxUnreadCount
		return
	}

	c.unreadCount += 1
}
func (c *Conversation) substractUnread(at time.Time) {
	c.updateUpdatedAt(at)

	if c.unreadCount <= minUnreadCount {
		c.unreadCount = minUnreadCount
		return
	}

	c.unreadCount -= 1
}
