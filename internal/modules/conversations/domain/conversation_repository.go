package domain

import (
	"context"
	"uuid"
)

type ConversationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	FindLastOpenByContactID(ctx context.Context, contactID uuid.UUID) (*Conversation, error)
	FindWithSpecificMessageByExternalID(ctx context.Context, externalMessageID string) (*Conversation, error)
	Save(ctx context.Context, conversation *Conversation) error
}
