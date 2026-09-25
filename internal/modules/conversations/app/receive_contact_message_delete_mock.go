package app

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockReceiveContactMessageDelete struct {
	mock.Mock
}

func (m *MockReceiveContactMessageDelete) Execute(ctx context.Context, input ReceiveContactMessageDeleteInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
