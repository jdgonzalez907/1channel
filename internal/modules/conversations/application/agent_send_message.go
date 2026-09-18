package application

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	AgentSendMessageInput struct {
		ConversationID uuid.UUID
		MessageID      uuid.UUID
		AgentID        uuid.UUID
		Text           string
		SentAt         time.Time
	}

	AgentSendMessage interface {
		Execute(ctx context.Context, input AgentSendMessageInput) error
	}

	agentSendMessage struct {
		conversationRepository domain.ConversationRepository
	}
)

func NewAgentSendMessage(conversationRepository domain.ConversationRepository) AgentSendMessage {
	return &agentSendMessage{conversationRepository}
}

func (uc *agentSendMessage) Execute(ctx context.Context, input AgentSendMessageInput) error {
	conversation, err := uc.conversationRepository.FindByID(ctx, input.ConversationID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.AgentSendMessage(
		input.MessageID,
		input.AgentID,
		input.Text,
		input.SentAt,
	)
	if err != nil {
		return uc.wrapError(err)
	}

	err = uc.conversationRepository.Save(ctx, conversation)
	if err != nil {
		return uc.wrapError(err)
	}

	return nil
}

func (uc *agentSendMessage) wrapError(err error) error {
	return errors.Join(errors.New("error sending agent message"), err)
}
