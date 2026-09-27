package domain

import (
	"context"
	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)

	user, _ := args.Get(0).(*User)
	return user, args.Error(1)
}

func (m *MockUserRepository) Save(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}
