package domain

import (
	"context"
	"uuid"
)

type PersonalInformationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*PersonalInformation, error)
	FindByIdentificationNumber(ctx context.Context, identificationNumber string) (*PersonalInformation, error)
	Save(ctx context.Context, personalInformation *PersonalInformation) error
}
