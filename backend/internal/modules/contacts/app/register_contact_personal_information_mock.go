package app

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

type MockRegisterContactPersonalInformation struct {
	mock.Mock
}

func (m *MockRegisterContactPersonalInformation) Execute(ctx context.Context, input RegisterContactPersonalInformationInput) (*domain.PersonalInformation, error) {
	args := m.Called(ctx, input)

	personalInformation, _ := args.Get(0).(*domain.PersonalInformation)
	return personalInformation, args.Error(1)
}
