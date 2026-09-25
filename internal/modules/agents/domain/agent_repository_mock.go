package domain

import (
	"context"

	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockAgentRepository struct {
	mock.Mock
}

func (m *MockAgentRepository) FindByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	args := m.Called(ctx, id)

	agent, _ := args.Get(0).(*Agent)
	return agent, args.Error(1)
}
