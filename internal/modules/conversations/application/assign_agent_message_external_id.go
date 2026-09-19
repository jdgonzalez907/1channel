package application

import (
	"context"
	"errors"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	AssignAgentMessageExternalIDInput struct {
		MessageID         uuid.UUID
		ExternalMessageID string
	}

	AssignAgentMessageExternalID interface {
		Execute(ctx context.Context, input AssignAgentMessageExternalIDInput) error
	}

	assignAgentMessageExternalID struct {
		conversationRepository domain.ConversationRepository
	}
)

func NewAssignAgentMessageExternalID(conversationRepository domain.ConversationRepository) AssignAgentMessageExternalID {
	return &assignAgentMessageExternalID{conversationRepository}
}

func (uc *assignAgentMessageExternalID) Execute(ctx context.Context, input AssignAgentMessageExternalIDInput) error {
	conversation, err := uc.conversationRepository.FindWithSpecificMessageByMessageID(ctx, input.MessageID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.AssignAgentMessageExternalID(input.MessageID, input.ExternalMessageID)
	if err != nil {
		return uc.wrapError(err)
	}

	err = uc.conversationRepository.Save(ctx, conversation)
	if err != nil {
		return uc.wrapError(err)
	}

	return nil
}

func (uc *assignAgentMessageExternalID) wrapError(err error) error {
	return errors.Join(errors.New("error assigning agent message external id"), err)
}
