package domain

type ConversationStatus string

const (
	Pending  ConversationStatus = "pending"
	Assigned ConversationStatus = "assigned"
	Expired  ConversationStatus = "expired"
	Resolved ConversationStatus = "resolved"
)

func NewConversationStatus(value string) (ConversationStatus, error) {
	return ConversationStatus(value), nil
}

func (s ConversationStatus) Value() string { return string(s) }
