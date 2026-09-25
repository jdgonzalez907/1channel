package app

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockResolveConversation struct {
	mock.Mock
}

func (m *MockResolveConversation) Execute(ctx context.Context, input ResolveConversationInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
