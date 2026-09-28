package domain

import "errors"

type ConversationStatus string

const (
	ConversationStatusPending  ConversationStatus = "pending"
	ConversationStatusAssigned ConversationStatus = "assigned"
	ConversationStatusExpired  ConversationStatus = "expired"
	ConversationStatusResolved ConversationStatus = "resolved"
)

var ErrConversationStatusInvalid = errors.New("conversation status is invalid")

func NewConversationStatus(s string) (ConversationStatus, error) {
	switch ConversationStatus(s) {
	case ConversationStatusPending, ConversationStatusAssigned, ConversationStatusExpired, ConversationStatusResolved:
		return ConversationStatus(s), nil
	default:
		return "", ErrConversationStatusInvalid
	}
}

func (s ConversationStatus) String() string {
	return string(s)
}
