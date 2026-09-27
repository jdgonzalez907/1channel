package app

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

type MockFindUserByID struct {
	mock.Mock
}

func (m *MockFindUserByID) Execute(ctx context.Context, input FindUserByIDInput) (*domain.User, error) {
	args := m.Called(ctx, input)

	user, _ := args.Get(0).(*domain.User)
	return user, args.Error(1)
}
