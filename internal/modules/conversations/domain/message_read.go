package conversations

import (
	"time"
)

type MessageRead struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
	ExternalID string    `json:"external_id"`
	ReadAt     time.Time `json:"read_at"`
}

func (e MessageRead) EventName() string { return "conversation.message_read" }

func NewMessageRead(id string, occurredAt time.Time, externalID string, readAt time.Time) MessageRead {
	return MessageRead{ID: id, OccurredAt: occurredAt, ExternalID: externalID, ReadAt: readAt}
}
