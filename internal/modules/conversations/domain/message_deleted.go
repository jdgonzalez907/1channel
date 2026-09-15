package conversations

import (
	"time"
)

type MessageDeleted struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
	ExternalID string    `json:"external_id"`
	DeletedAt  time.Time `json:"deleted_at"`
}

func (e MessageDeleted) EventName() string { return "conversation.message_deleted" }

func NewMessageDeleted(id string, occurredAt time.Time, externalID string, deletedAt time.Time) MessageDeleted {
	return MessageDeleted{ID: id, OccurredAt: occurredAt, ExternalID: externalID, DeletedAt: deletedAt}
}
