package domain

import (
	"context"
	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockPersonalInformationRepository struct {
	mock.Mock
}

func (m *MockPersonalInformationRepository) FindByID(ctx context.Context, id uuid.UUID) (*PersonalInformation, error) {
	args := m.Called(ctx, id)

	personalInformation, _ := args.Get(0).(*PersonalInformation)
	return personalInformation, args.Error(1)
}

func (m *MockPersonalInformationRepository) FindByIdentificationNumber(ctx context.Context, identificationNumber string) (*PersonalInformation, error) {
	args := m.Called(ctx, identificationNumber)

	personalInformation, _ := args.Get(0).(*PersonalInformation)
	return personalInformation, args.Error(1)
}

func (m *MockPersonalInformationRepository) Save(ctx context.Context, personalInformation *PersonalInformation) error {
	args := m.Called(ctx, personalInformation)

	return args.Error(0)
}
