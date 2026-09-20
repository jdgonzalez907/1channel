package domain

import (
	"context"
)

type ContactRepository interface {
	FindByExternalContactID(ctx context.Context, externalContactID string) (*Contact, error)
	Save(ctx context.Context, contact *Contact) error
}
