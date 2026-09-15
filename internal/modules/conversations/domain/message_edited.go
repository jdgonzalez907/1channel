package conversations

import (
	"time"
)

type MessageEdited struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
	ExternalID string    `json:"external_id"`
	Text       string    `json:"text"`
	EditedAt   time.Time `json:"edited_at"`
}

func (e MessageEdited) EventName() string { return "conversation.message_edited" }

func NewMessageEdited(id string, occurredAt time.Time, externalID string, text string, editedAt time.Time) MessageEdited {
	return MessageEdited{ID: id, OccurredAt: occurredAt, ExternalID: externalID, Text: text, EditedAt: editedAt}
}
