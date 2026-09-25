package app

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockReceiveContactMessageEdit struct {
	mock.Mock
}

func (m *MockReceiveContactMessageEdit) Execute(ctx context.Context, input ReceiveContactMessageEditInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
