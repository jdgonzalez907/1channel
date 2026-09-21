package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/jdgonzalez907/1channel/internal/postgres/sqlc"
)

type contactRepository struct {
	queries *sqlc.Queries
}

func NewContactRepository(pool *pgxpool.Pool) domain.ContactRepository {
	return &contactRepository{sqlc.New(pool)}
}

func (r *contactRepository) FindByExternalContactID(ctx context.Context, externalContactID string) (*domain.Contact, error) {
	row, err := r.queries.FindContactByExternalID(ctx, externalContactID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return domain.NewContact(pgUUIDToUUID(row.Contact.ID), row.Contact.ExternalContactID, row.Contact.CreatedAt.Time)
}

func (r *contactRepository) Save(ctx context.Context, contact *domain.Contact) error {
	return r.queries.UpsertContact(ctx, sqlc.UpsertContactParams{
		ID:                pgUUIDFromUUID(contact.ID()),
		ExternalContactID: contact.ExternalContactID(),
		CreatedAt:         toTimestamptz(contact.CreatedAt()),
	})
}
