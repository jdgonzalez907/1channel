package domain

import "context"

type AgentMessageSender interface {
	Send(ctx context.Context, to string, message *Message) (externalMessageID string, err error)
}
