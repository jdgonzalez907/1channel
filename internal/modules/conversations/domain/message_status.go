package domain

type MessageStatus string

const (
	Registered MessageStatus = "registered"
	Sent       MessageStatus = "sent"
	Delivered  MessageStatus = "delivered"
	Read       MessageStatus = "read"
	Deleted    MessageStatus = "deleted"
)

var (
	MessageStatusRank map[MessageStatus]int8 = map[MessageStatus]int8{
		Registered: 1,
		Sent:       2,
		Delivered:  3,
		Read:       4,
		Deleted:    5,
	}
)

func NewMessageStatus(value string) (MessageStatus, error) {
	return MessageStatus(value), nil
}

func (s MessageStatus) Value() string { return string(s) }
