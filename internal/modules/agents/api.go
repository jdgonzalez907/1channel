package agents

import (
	"context"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/app"
)

type AgentsAPI interface {
	FindAgentByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

type agentsAPI struct {
	findAgentByID app.FindAgentByID
}

func NewAgentsAPI(findAgentByID app.FindAgentByID) AgentsAPI {
	return &agentsAPI{findAgentByID: findAgentByID}
}

func (a *agentsAPI) FindAgentByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	agent, err := a.findAgentByID.Execute(ctx, app.FindAgentByIDInput{ID: id})
	if err != nil {
		return uuid.Nil(), err
	}

	return agent.ID(), nil
}
