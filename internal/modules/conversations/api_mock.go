package conversations

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockConversationsAPI struct {
	mock.Mock
}

func NewMockConversationsAPI() *MockConversationsAPI {
	return &MockConversationsAPI{}
}

func (m *MockConversationsAPI) ReceiveContactMessage(ctx context.Context, input ReceiveContactMessageInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
