package app

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

type MockFindContactByID struct {
	mock.Mock
}

func (m *MockFindContactByID) Execute(ctx context.Context, input FindContactByIDInput) (*domain.Contact, error) {
	args := m.Called(ctx, input)

	contact, _ := args.Get(0).(*domain.Contact)
	return contact, args.Error(1)
}
