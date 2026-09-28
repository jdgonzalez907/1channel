package domain

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAgentMessageSender struct {
	mock.Mock
}

func (m *MockAgentMessageSender) Send(ctx context.Context, to string, message *Message) (string, error) {
	args := m.Called(ctx, to, message)

	return args.String(0), args.Error(1)
}
