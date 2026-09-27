package app

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

type MockCreateUser struct {
	mock.Mock
}

func (m *MockCreateUser) Execute(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)

	user, _ := args.Get(0).(*domain.User)
	return user, args.Error(1)
}
