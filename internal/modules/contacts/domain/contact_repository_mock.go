package domain

import (
	"context"
	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*Contact, error) {
	args := m.Called(ctx, id)

	contact, _ := args.Get(0).(*Contact)
	return contact, args.Error(1)
}

func (m *MockContactRepository) FindByExternalContactID(ctx context.Context, externalContactID string) (*Contact, error) {
	args := m.Called(ctx, externalContactID)

	contact, _ := args.Get(0).(*Contact)
	return contact, args.Error(1)
}

func (m *MockContactRepository) Save(ctx context.Context, contact *Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}
