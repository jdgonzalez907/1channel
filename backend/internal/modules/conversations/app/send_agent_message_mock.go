package app

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockSendAgentMessage struct {
	mock.Mock
}

func (m *MockSendAgentMessage) Execute(ctx context.Context, input SendAgentMessageInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
