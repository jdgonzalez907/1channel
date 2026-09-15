package conversations

import (
	"time"
)

type MessageDelivered struct {
	ID          string    `json:"id"`
	OccurredAt  time.Time `json:"occurred_at"`
	ExternalID  string    `json:"external_id"`
	DeliveredAt time.Time `json:"delivered_at"`
}

func (e MessageDelivered) EventName() string { return "conversation.message_delivered" }

func NewMessageDelivered(id string, occurredAt time.Time, externalID string, deliveredAt time.Time) MessageDelivered {
	return MessageDelivered{ID: id, OccurredAt: occurredAt, ExternalID: externalID, DeliveredAt: deliveredAt}
}
