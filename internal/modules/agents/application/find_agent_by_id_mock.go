package application

import (
	"context"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/stretchr/testify/mock"
)

type MockFindAgentByID struct {
	FindAgentByID
	mock.Mock
}

func (m *MockFindAgentByID) Execute(ctx context.Context, input FindAgentByIDInput) (*domain.Agent, error) {
	args := m.Called(ctx, input)
	return agentFromApplication(args), args.Error(1)
}

func agentFromApplication(args mock.Arguments) *domain.Agent {
	value := args.Get(0)
	if value == nil {
		return nil
	}

	return value.(*domain.Agent)
}
