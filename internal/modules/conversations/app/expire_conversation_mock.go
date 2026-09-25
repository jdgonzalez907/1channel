package app

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockExpireConversation struct {
	mock.Mock
}

func (m *MockExpireConversation) Execute(ctx context.Context, input ExpireConversationInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
