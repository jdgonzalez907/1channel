package app

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAgentReadConversation struct {
	mock.Mock
}

func (m *MockAgentReadConversation) Execute(ctx context.Context, input AgentReadConversationInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
