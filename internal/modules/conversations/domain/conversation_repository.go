package domain

import (
	"context"

	"uuid"
)

type ConversationRepository interface {
	FindWithoutMessages(ctx context.Context, id uuid.UUID) (*Conversation, error)
	FindWithContactUnreadMessagesByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	FindWithMessageByExternalID(ctx context.Context, externalMessageID string) (*Conversation, error)
	FindOpenByContactID(ctx context.Context, contactID uuid.UUID) (*Conversation, error)
	Save(ctx context.Context, conversation *Conversation) error
}
