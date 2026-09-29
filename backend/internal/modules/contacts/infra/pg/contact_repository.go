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

type postgresContactRepository struct {
	db *pgdb.DB
}

func NewContactRepository(db *pgdb.DB) domain.ContactRepository {
	return &postgresContactRepository{db: db}
}

func (r *postgresContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	row, err := r.db.Queries.FindContactByID(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toContact(row), nil
}

func (r *postgresContactRepository) FindByExternalContactID(ctx context.Context, externalContactID string) (*domain.Contact, error) {
	row, err := r.db.Queries.FindContactByExternalContactID(ctx, externalContactID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toContact(row), nil
}

func (r *postgresContactRepository) Save(ctx context.Context, contact *domain.Contact) error {
	return r.db.Queries.UpsertContact(ctx, sqlc.UpsertContactParams{
		ID:                    pgdb.UUID(contact.ID()),
		ExternalContactID:     contact.ExternalContactID(),
		DisplayName:           contact.DisplayName(),
		PersonalInformationID: pgdb.UUIDPtr(contact.PersonalInformationID()),
		CreatedAt:             pgdb.Timestamp(contact.CreatedAt()),
	})
}

func toContact(row sqlc.Contact) *domain.Contact {
	return domain.RehydrateContact(
		pgdb.FromUUID(row.ID),
		row.ExternalContactID,
		row.DisplayName,
		pgdb.FromUUIDPtr(row.PersonalInformationID),
		pgdb.FromTimestamp(row.CreatedAt),
	)
}
