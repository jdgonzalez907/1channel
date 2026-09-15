package conversations

import "context"

type ConversationRepository interface {
	FindByExternalIDs(ctx context.Context, messageExternalID []string) ([]*Conversation, error)
	Save(ctx context.Context, conversation *Conversation) error
}
