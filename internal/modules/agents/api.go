package agents

import (
	"context"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/application"
)

type AgentsAPI interface {
	FindAgentByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

type agentsAPI struct {
	findAgentByID application.FindAgentByID
}

func NewAgentsAPI(findAgentByID application.FindAgentByID) AgentsAPI {
	return &agentsAPI{findAgentByID}
}

func (a *agentsAPI) FindAgentByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	agent, err := a.findAgentByID.Execute(ctx, application.FindAgentByIDInput{ID: id})
	if err != nil {
		return uuid.Nil(), err
	}

	return agent.ID(), nil
}
