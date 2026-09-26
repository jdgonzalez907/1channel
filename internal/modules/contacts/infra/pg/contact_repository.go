package pg

import (
	"context"
	"errors"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

type contactRepository struct {
	db *pgdb.DB
}

func NewContactRepository(db *pgdb.DB) domain.ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	row, err := r.db.Queries.FindContactByID(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toContact(row)
}

func (r *contactRepository) FindByExternalContactID(ctx context.Context, externalContactID string) (*domain.Contact, error) {
	row, err := r.db.Queries.FindContactByExternalContactID(ctx, externalContactID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toContact(row)
}

func (r *contactRepository) Save(ctx context.Context, contact *domain.Contact) error {
	return r.db.Queries.UpsertContact(ctx, sqlc.UpsertContactParams{
		ID:                pgdb.UUID(contact.ID()),
		ExternalContactID: contact.ExternalContactID(),
		CreatedAt:         pgdb.Timestamp(contact.CreatedAt()),
	})
}

func toContact(row sqlc.Contact) (*domain.Contact, error) {
	return domain.NewContact(pgdb.FromUUID(row.ID), row.ExternalContactID, pgdb.FromTimestamp(row.CreatedAt))
}
