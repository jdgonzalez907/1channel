package domain

import "errors"

type MessageStatus string

const (
	MessageStatusSent    MessageStatus = "sent"
	MessageStatusRead    MessageStatus = "read"
	MessageStatusDeleted MessageStatus = "deleted"
)

var ErrMessageStatusInvalid = errors.New("message status is invalid")

func NewMessageStatus(s string) (MessageStatus, error) {
	switch MessageStatus(s) {
	case MessageStatusSent, MessageStatusRead, MessageStatusDeleted:
		return MessageStatus(s), nil
	default:
		return "", ErrMessageStatusInvalid
	}
}

func (s MessageStatus) String() string {
	return string(s)
}
