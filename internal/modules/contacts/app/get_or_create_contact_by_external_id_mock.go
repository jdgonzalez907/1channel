package app

import (
	"context"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/stretchr/testify/mock"
)

type MockGetOrCreateContactByExternalID struct {
	mock.Mock
}

func (m *MockGetOrCreateContactByExternalID) Execute(ctx context.Context, input GetOrCreateContactByExternalIDInput) (*domain.Contact, error) {
	args := m.Called(ctx, input)

	contact, _ := args.Get(0).(*domain.Contact)
	return contact, args.Error(1)
}
