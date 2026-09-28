package app

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockReceiveContactMessageRead struct {
	mock.Mock
}

func (m *MockReceiveContactMessageRead) Execute(ctx context.Context, input ReceiveContactMessageReadInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
