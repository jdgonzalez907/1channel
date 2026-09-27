package users

import (
	"context"
	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockUsersAPI struct {
	mock.Mock
}

func (m *MockUsersAPI) FindUserByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, id)

	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUsersAPI) CreateUser(ctx context.Context, input CreateUserInput) (uuid.UUID, error) {
	args := m.Called(ctx, input)

	return args.Get(0).(uuid.UUID), args.Error(1)
}
