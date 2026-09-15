package conversations

import (
	"time"
)

type MessageReceived struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
	ExternalID string    `json:"external_id"`
	Text       string    `json:"text"`
	FromAgent  bool      `json:"from_agent"`
}

func (e MessageReceived) EventName() string { return "conversation.message_received" }

func NewMessageReceived(id string, occurredAt time.Time, externalID string, text string, fromAgent bool) MessageReceived {
	return MessageReceived{ID: id, OccurredAt: occurredAt, ExternalID: externalID, Text: text, FromAgent: fromAgent}
}
