package domain

import (
	"time"
)

type ConversationStatus string

const (
	Unassigned  ConversationStatus = "unassigned"
	Assigned    ConversationStatus = "assigned"
	Transferred ConversationStatus = "transferred"
	Finished    ConversationStatus = "finished"
	Abandoned   ConversationStatus = "abandoned"
)

type Conversation struct {
	id                string
	platformContactID int64
	agentID           *int64
	status            ConversationStatus
	unreadCount       int8
	existingMessages  map[string]*Message
	addedMessages     map[string]*Message
	updatedMessages   map[string]*Message
	removedMessages   map[string]bool
	createdAt         time.Time
	updatedAt         *time.Time
	closedAt          *time.Time
}

func (c *Conversation) ID() string                            { return c.id }
func (c *Conversation) PlatformContactID() int64              { return c.platformContactID }
func (c *Conversation) AgentID() *int64                       { return c.agentID }
func (c *Conversation) Status() ConversationStatus            { return c.status }
func (c *Conversation) UnreadCount() int8                     { return c.unreadCount }
func (c *Conversation) ExistingMessages() map[string]*Message { return c.existingMessages }
func (c *Conversation) AddedMessages() map[string]*Message    { return c.addedMessages }
func (c *Conversation) UpdatedMessages() map[string]*Message  { return c.updatedMessages }
func (c *Conversation) RemovedMessages() map[string]bool      { return c.removedMessages }
func (c *Conversation) CreatedAt() time.Time                  { return c.createdAt }
func (c *Conversation) UpdatedAt() *time.Time                 { return c.updatedAt }
func (c *Conversation) ClosedAt() *time.Time                  { return c.closedAt }

func NewConversation(
	id string,
	platformContactID int64,
	agentID *int64,
	status ConversationStatus,
	unreadCount int8,
	existingMessages,
	addedMessages,
	updatedMessages map[string]*Message,
	removedMessages map[string]bool,
	createdAt time.Time,
	updatedAt, closedAt *time.Time,
) *Conversation {
	return &Conversation{
		id,
		platformContactID,
		agentID,
		status,
		unreadCount,
		existingMessages,
		addedMessages,
		updatedMessages,
		removedMessages,
		createdAt,
		updatedAt,
		closedAt,
	}
}
