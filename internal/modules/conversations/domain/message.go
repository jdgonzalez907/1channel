package conversations

import "time"

type MessageStatus string

const (
	Sent      MessageStatus = "sent"
	Delivered MessageStatus = "delivered"
	Read      MessageStatus = "read"
	Deleted   MessageStatus = "deleted"
)

func NewMessageStatus(value string) (MessageStatus, error) {
	return MessageStatus(value), nil
}

func (s MessageStatus) Value() string { return string(s) }

type Message struct {
	id        string
	text      string
	status    MessageStatus
	fromAgent bool
	createdAt time.Time
	updatedAt *time.Time
	deletedAt *time.Time
	readAt    *time.Time
}

func NewMessage(
	id string,
	text string,
	status MessageStatus,
	fromAgent bool,
	createdAt time.Time,
	updatedAt *time.Time,
	deletedAt *time.Time,
	readAt *time.Time,
) (*Message, error) {
	return &Message{
		id,
		text,
		status,
		fromAgent,
		createdAt,
		updatedAt,
		deletedAt,
		readAt,
	}, nil
}

func (m *Message) ID() string            { return m.id }
func (m *Message) Text() string          { return m.text }
func (m *Message) Status() MessageStatus { return m.status }
func (m *Message) FromAgent() bool       { return m.fromAgent }
func (m *Message) CreatedAt() time.Time  { return m.createdAt }
func (m *Message) UpdatedAt() *time.Time { return m.updatedAt }
func (m *Message) DeletedAt() *time.Time { return m.deletedAt }
func (m *Message) ReadAt() *time.Time    { return m.readAt }

type MessageDTO struct {
	ID         string        `json:"id"`
	ExternalID string        `json:"external_id"`
	Text       string        `json:"text"`
	Status     MessageStatus `json:"status"`
	FromAgent  bool          `json:"from_agent"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  *time.Time    `json:"updated_at,omitempty"`
	DeletedAt  *time.Time    `json:"deleted_at,omitempty"`
	ReadAt     *time.Time    `json:"read_at,omitempty"`
}
