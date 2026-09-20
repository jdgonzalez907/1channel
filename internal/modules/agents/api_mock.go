package agents

import (
	"context"

	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockAgentsAPI struct {
	AgentsAPI
	mock.Mock
}

func (m *MockAgentsAPI) FindAgentByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(uuid.UUID), args.Error(1)
}
