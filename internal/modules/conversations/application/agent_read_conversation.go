package application

import (
	"context"
	"errors"
	"time"
	"uuid"

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
	}
)

func NewAgentReadConversation(conversationRepository domain.ConversationRepository) AgentReadConversation {
	return &agentReadConversation{conversationRepository}
}

func (uc *agentReadConversation) Execute(ctx context.Context, input AgentReadConversationInput) error {
	conversation, err := uc.conversationRepository.FindByID(ctx, input.ConversationID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.AgentReadConversation(input.AgentID, input.ReadAt)
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
