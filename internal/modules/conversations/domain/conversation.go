package conversations

import "time"

type ConversationStatus string

const (
	Opened     ConversationStatus = "opened"
	InProgress ConversationStatus = "in_progress"
	Finished   ConversationStatus = "finished"
	Closed     ConversationStatus = "closed"
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
	agentID         string
	contactID       string
	createdAt       time.Time
	updatedAt       *time.Time
	finishedAt      *time.Time
	deletedAt       *time.Time
}

func NewConversation(
	id string,
	status ConversationStatus,
	messages map[string]*Message,
	foundMessages map[string]*Message,
	addedMessages map[string]*Message,
	updatedMessages map[string]*Message,
	deletedMessages map[string]*Message,
	unreadCount int,
	agentID string,
	contactID string,
	createdAt time.Time,
	updatedAt *time.Time,
	finishedAt *time.Time,
	deletedAt *time.Time,
) (*Conversation, error) {
	return &Conversation{
		id:              id,
		status:          status,
		foundMessages:   foundMessages,
		addedMessages:   addedMessages,
		updatedMessages: updatedMessages,
		deletedMessages: deletedMessages,
		unreadCount:     unreadCount,
		agentID:         agentID,
		contactID:       contactID,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		finishedAt:      finishedAt,
		deletedAt:       deletedAt,
	}, nil
}

func (c *Conversation) ID() string                           { return c.id }
func (c *Conversation) Status() ConversationStatus           { return c.status }
func (c *Conversation) FoundMessages() map[string]*Message   { return c.foundMessages }
func (c *Conversation) AddedMessages() map[string]*Message   { return c.addedMessages }
func (c *Conversation) UpdatedMessages() map[string]*Message { return c.updatedMessages }
func (c *Conversation) DeletedMessages() map[string]*Message { return c.deletedMessages }
func (c *Conversation) UnreadCount() int                     { return c.unreadCount }
func (c *Conversation) AgentID() string                      { return c.agentID }
func (c *Conversation) ContactID() string                    { return c.contactID }
func (c *Conversation) CreatedAt() time.Time                 { return c.createdAt }
func (c *Conversation) UpdatedAt() *time.Time                { return c.updatedAt }
func (c *Conversation) FinishedAt() *time.Time               { return c.finishedAt }
func (c *Conversation) DeletedAt() *time.Time                { return c.deletedAt }

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
