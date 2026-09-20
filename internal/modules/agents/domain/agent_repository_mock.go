package domain

import (
	"context"
	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockAgentRepository struct {
	AgentRepository
	mock.Mock
}

func (m *MockAgentRepository) FindByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	args := m.Called(ctx, id)
	return agentFrom(args), args.Error(1)
}

func agentFrom(args mock.Arguments) *Agent {
	value := args.Get(0)
	if value == nil {
		return nil
	}

	return value.(*Agent)
}
