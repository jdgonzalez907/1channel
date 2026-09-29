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

type postgresPersonalInformationRepository struct {
	db *pgdb.DB
}

func NewPersonalInformationRepository(db *pgdb.DB) domain.PersonalInformationRepository {
	return &postgresPersonalInformationRepository{db: db}
}

func (r *postgresPersonalInformationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.PersonalInformation, error) {
	row, err := r.db.Queries.FindPersonalInformationByID(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toPersonalInformation(row), nil
}

func (r *postgresPersonalInformationRepository) FindByIdentificationNumber(ctx context.Context, identificationNumber string) (*domain.PersonalInformation, error) {
	row, err := r.db.Queries.FindPersonalInformationByIdentificationNumber(ctx, identificationNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toPersonalInformation(row), nil
}

func (r *postgresPersonalInformationRepository) Save(ctx context.Context, personalInformation *domain.PersonalInformation) error {
	row, err := r.db.Queries.UpsertPersonalInformation(ctx, sqlc.UpsertPersonalInformationParams{
		ID:                   pgdb.UUID(personalInformation.ID()),
		IdentificationNumber: personalInformation.IdentificationNumber(),
		FirstName:            personalInformation.FirstName(),
		LastName:             personalInformation.LastName(),
		PhoneNumber:          personalInformation.PhoneNumber(),
		Email:                personalInformation.Email(),
		Address:              personalInformation.Address(),
		CreatedAt:            pgdb.Timestamp(personalInformation.CreatedAt()),
		UpdatedAt:            pgdb.Timestamp(personalInformation.UpdatedAt()),
	})
	if err != nil {
		return err
	}

	personalInformation.AssignID(pgdb.FromUUID(row.ID))

	return nil
}

func toPersonalInformation(row sqlc.PersonalInformation) *domain.PersonalInformation {
	return domain.RehydratePersonalInformation(
		pgdb.FromUUID(row.ID),
		row.IdentificationNumber,
		row.FirstName,
		row.LastName,
		row.PhoneNumber,
		row.Email,
		row.Address,
		pgdb.FromTimestamp(row.CreatedAt),
		pgdb.FromTimestamp(row.UpdatedAt),
	)
}
