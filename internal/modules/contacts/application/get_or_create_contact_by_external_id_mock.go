package application

import (
	"context"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/stretchr/testify/mock"
)

type MockGetOrCreateContactByExternalID struct {
	GetOrCreateContactByExternalID
	mock.Mock
}

func (m *MockGetOrCreateContactByExternalID) Execute(ctx context.Context, input GetOrCreateContactByExternalIDInput) (*domain.Contact, error) {
	args := m.Called(ctx, input)
	return contactFromApplication(args), args.Error(1)
}

func contactFromApplication(args mock.Arguments) *domain.Contact {
	value := args.Get(0)
	if value == nil {
		return nil
	}

	return value.(*domain.Contact)
}
