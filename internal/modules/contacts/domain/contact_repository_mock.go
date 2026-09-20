package domain

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockContactRepository struct {
	ContactRepository
	mock.Mock
}

func (m *MockContactRepository) FindByExternalContactID(ctx context.Context, externalContactID string) (*Contact, error) {
	args := m.Called(ctx, externalContactID)
	return contactFrom(args), args.Error(1)
}

func (m *MockContactRepository) Save(ctx context.Context, contact *Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func contactFrom(args mock.Arguments) *Contact {
	value := args.Get(0)
	if value == nil {
		return nil
	}

	return value.(*Contact)
}
