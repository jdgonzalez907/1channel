package application

import (
	"context"
	"errors"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"uuid"
)

type (
	FindAgentByIDInput struct {
		ID uuid.UUID
	}

	FindAgentByID interface {
		Execute(ctx context.Context, input FindAgentByIDInput) (*domain.Agent, error)
	}

	findAgentByID struct {
		agentRepository domain.AgentRepository
	}
)

func NewFindAgentByID(agentRepository domain.AgentRepository) FindAgentByID {
	return &findAgentByID{agentRepository}
}

func (uc *findAgentByID) Execute(ctx context.Context, input FindAgentByIDInput) (*domain.Agent, error) {
	agent, err := uc.agentRepository.FindByID(ctx, input.ID)
	if err != nil {
		return nil, uc.wrapError(err)
	}

	if agent == nil {
		return nil, uc.wrapError(domain.ErrAgentNotFound)
	}

	return agent, nil
}

func (uc *findAgentByID) wrapError(err error) error {
	return errors.Join(errors.New("error finding agent by id"), err)
}
