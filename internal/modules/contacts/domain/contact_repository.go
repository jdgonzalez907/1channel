package domain

import (
	"context"
	"uuid"
)

type ContactRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)
	FindByExternalContactID(ctx context.Context, externalContactID string) (*Contact, error)
	Save(ctx context.Context, contact *Contact) error
}
