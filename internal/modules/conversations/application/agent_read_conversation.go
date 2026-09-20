package application

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	AgentReadConversationInput struct {
		ConversationID uuid.UUID
		AgentID        uuid.UUID
		ReadAt         time.Time
	}

	AgentReadConversation interface {
		Execute(ctx context.Context, input AgentReadConversationInput) error
	}

	agentReadConversation struct {
		conversationRepository domain.ConversationRepository
		agentsAPI              agents.AgentsAPI
	}
)

func NewAgentReadConversation(conversationRepository domain.ConversationRepository, agentsAPI agents.AgentsAPI) AgentReadConversation {
	return &agentReadConversation{conversationRepository, agentsAPI}
}

func (uc *agentReadConversation) Execute(ctx context.Context, input AgentReadConversationInput) error {
	agentID, err := uc.agentsAPI.FindAgentByID(ctx, input.AgentID)
	if err != nil {
		return uc.wrapError(err)
	}

	conversation, err := uc.conversationRepository.FindByID(ctx, input.ConversationID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.AgentReadConversation(agentID, input.ReadAt)
	if err != nil {
		if errors.Is(err, domain.ErrConversationNotAssigned) {
			return nil
		}

		return uc.wrapError(err)
	}

	err = uc.conversationRepository.Save(ctx, conversation)
	if err != nil {
		return uc.wrapError(err)
	}

	return nil
}

func (uc *agentReadConversation) wrapError(err error) error {
	return errors.Join(errors.New("error agent reading conversation"), err)
}
