package domain

import (
	"context"
)

type MessageSender interface {
	Send(ctx context.Context, to string, message *Message) (externalMessageID string, err error)
}
