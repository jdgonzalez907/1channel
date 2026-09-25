package app

import (
	"context"
	"errors"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
)

var ErrFindingAgentByID = errors.New("finding agent by id")

type FindAgentByIDInput struct {
	ID uuid.UUID
}

type FindAgentByID interface {
	Execute(ctx context.Context, input FindAgentByIDInput) (*domain.Agent, error)
}

type findAgentByID struct {
	agentRepository domain.AgentRepository
}

func NewFindAgentByID(agentRepository domain.AgentRepository) FindAgentByID {
	return &findAgentByID{agentRepository: agentRepository}
}

func (uc *findAgentByID) Execute(ctx context.Context, input FindAgentByIDInput) (*domain.Agent, error) {
	agent, err := uc.agentRepository.FindByID(ctx, input.ID)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	if agent == nil {
		return nil, uc.joinErr(domain.ErrAgentNotFound)
	}

	return agent, nil
}

func (uc *findAgentByID) joinErr(errs ...error) error {
	return errors.Join(ErrFindingAgentByID, errors.Join(errs...))
}
