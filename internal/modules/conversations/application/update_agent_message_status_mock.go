package application

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUpdateAgentMessageStatus struct {
	mock.Mock
}

func NewMockUpdateAgentMessageStatus() *MockUpdateAgentMessageStatus {
	return &MockUpdateAgentMessageStatus{}
}

func (m *MockUpdateAgentMessageStatus) Execute(ctx context.Context, input UpdateAgentMessageStatusInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
