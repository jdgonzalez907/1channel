package domain

import "errors"

type MessageType string

const (
	MessageTypeText MessageType = "text"
)

var ErrMessageTypeInvalid = errors.New("message type is invalid")

func NewMessageType(s string) (MessageType, error) {
	switch MessageType(s) {
	case MessageTypeText:
		return MessageType(s), nil
	default:
		return "", ErrMessageTypeInvalid
	}
}

func (t MessageType) String() string {
	return string(t)
}
