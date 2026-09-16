package conversations

import (
	"errors"
	"time"
)

type ConversationStatus string

const (
	Pending  ConversationStatus = "pending"
	Assigned ConversationStatus = "assigned"
	Expired  ConversationStatus = "expired"
	Resolved ConversationStatus = "resolved"
)

var (
	ErrInvalidState         = errors.New("invalid state")
	ErrConversationExpired  = errors.New("conversation expired")
	ErrConversationResolved = errors.New("conversation resolved")
	ErrUnauthorizedAgent    = errors.New("unauthorized agent")
	ErrMessageNotFound      = errors.New("message not found")
)

func NewConversationStatus(value string) (ConversationStatus, error) {
	return ConversationStatus(value), nil
}

func (s ConversationStatus) Value() string { return string(s) }

type Conversation struct {
	id              string
	status          ConversationStatus
	foundMessages   map[string]*Message
	addedMessages   map[string]*Message
	updatedMessages map[string]*Message
	deletedMessages map[string]*Message
	unreadCount     int
	agentID         *string
	contactID       string
	createdAt       time.Time
	updatedAt       *time.Time
	finishedAt      *time.Time
	deletedAt       *time.Time
}

func NewConversation(
	id string,
	status ConversationStatus,
	foundMessages map[string]*Message,
	addedMessages map[string]*Message,
	updatedMessages map[string]*Message,
	deletedMessages map[string]*Message,
	unreadCount int,
	agentID *string,
	contactID string,
	createdAt time.Time,
	updatedAt *time.Time,
	finishedAt *time.Time,
	deletedAt *time.Time,
) (*Conversation, error) {
	return &Conversation{
		id,
		status,
		foundMessages,
		addedMessages,
		updatedMessages,
		deletedMessages,
		unreadCount,
		agentID,
		contactID,
		createdAt,
		updatedAt,
		finishedAt,
		deletedAt,
	}, nil
}

func (c *Conversation) ID() string                           { return c.id }
func (c *Conversation) Status() ConversationStatus           { return c.status }
func (c *Conversation) FoundMessages() map[string]*Message   { return c.foundMessages }
func (c *Conversation) AddedMessages() map[string]*Message   { return c.addedMessages }
func (c *Conversation) UpdatedMessages() map[string]*Message { return c.updatedMessages }
func (c *Conversation) DeletedMessages() map[string]*Message { return c.deletedMessages }
func (c *Conversation) UnreadCount() int                     { return c.unreadCount }
func (c *Conversation) AgentID() *string                     { return c.agentID }
func (c *Conversation) ContactID() string                    { return c.contactID }
func (c *Conversation) CreatedAt() time.Time                 { return c.createdAt }
func (c *Conversation) UpdatedAt() *time.Time                { return c.updatedAt }
func (c *Conversation) FinishedAt() *time.Time               { return c.finishedAt }
func (c *Conversation) DeletedAt() *time.Time                { return c.deletedAt }
func (c *Conversation) CanAddMessage() bool {
	if c.status == Expired || c.status == Resolved {
		return false
	}
	return true
}
func (c *Conversation) AddMessage(messageID, text string, contactID *string, receivedAt time.Time) error {
	_, ok := c.foundMessages[messageID]
	if ok {
		return nil
	}

	fromAgent := contactID == nil
	var status MessageStatus
	if !fromAgent {
		status = Delivered
	} else {
		status = Sent
	}

	newMessage, err := NewMessage(messageID, text, status, fromAgent, receivedAt, nil, nil, nil)
	if err != nil {
		return err
	}

	if fromAgent {
		c.status = Assigned
	}

	c.addedMessages[messageID] = newMessage
	c.foundMessages[messageID] = newMessage
	c.updatedAt = &receivedAt

	return nil
}
func (c *Conversation) UpdateMessage(messageID, text string, editedAt time.Time) error {
	found, ok := c.foundMessages[messageID]
	if !ok {
		return ErrMessageNotFound
	}

	if found.updatedAt != nil && found.updatedAt.UTC().Equal(editedAt) {
		return nil
	}

	found.text = text
	found.updatedAt = &editedAt

	c.updatedMessages[messageID] = found
	c.foundMessages[messageID] = found
	c.updatedAt = &editedAt

	return nil
}
func (c *Conversation) DeleteMessage(messageID string, deletedAt time.Time) error {
	found, ok := c.foundMessages[messageID]
	if !ok {
		return ErrMessageNotFound
	}

	if found.deletedAt != nil {
		return nil
	}

	found.deletedAt = &deletedAt
	found.status = Deleted

	c.deletedMessages[messageID] = found
	c.foundMessages[messageID] = found
	c.updatedAt = &deletedAt

	return nil
}
func (c *Conversation) MarkMessageRead(messageID string, readAt time.Time) error {
	found, ok := c.foundMessages[messageID]
	if !ok {
		return ErrMessageNotFound
	}

	if found.readAt != nil && found.readAt.UTC().Equal(readAt) {
		return nil
	}

	found.readAt = &readAt
	found.status = Read

	c.updatedMessages[messageID] = found
	c.foundMessages[messageID] = found
	c.unreadCount = 0
	c.updatedAt = &readAt

	return nil
}

type ConversationDTO struct {
	ID          string             `json:"id"`
	Status      ConversationStatus `json:"status"`
	Messages    []MessageDTO       `json:"messages"`
	UnreadCount int                `json:"unread_count"`
	AgentID     string             `json:"agent_id"`
	ContactID   string             `json:"contact_id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   *time.Time         `json:"updated_at,omitempty"`
	FinishedAt  *time.Time         `json:"finished_at,omitempty"`
	DeletedAt   *time.Time         `json:"deleted_at,omitempty"`
}
