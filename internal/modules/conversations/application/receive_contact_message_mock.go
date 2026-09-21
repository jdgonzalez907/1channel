package application

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockReceiveContactMessage struct {
	mock.Mock
}

func NewMockReceiveContactMessage() *MockReceiveContactMessage {
	return &MockReceiveContactMessage{}
}

func (m *MockReceiveContactMessage) Execute(ctx context.Context, input ReceiveContactMessageInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
