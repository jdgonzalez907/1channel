package app

import (
	"context"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/stretchr/testify/mock"
)

type MockFindAgentByID struct {
	mock.Mock
}

func (m *MockFindAgentByID) Execute(ctx context.Context, input FindAgentByIDInput) (*domain.Agent, error) {
	args := m.Called(ctx, input)

	agent, _ := args.Get(0).(*domain.Agent)
	return agent, args.Error(1)
}
