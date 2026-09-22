package application

import (
	"context"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/stretchr/testify/mock"
)

type MockFindContactByID struct {
	FindContactByID
	mock.Mock
}

func (m *MockFindContactByID) Execute(ctx context.Context, input FindContactByIDInput) (*domain.Contact, error) {
	args := m.Called(ctx, input)
	return contactFromApplication(args), args.Error(1)
}
