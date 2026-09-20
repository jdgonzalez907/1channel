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
		agentsAPI              agents.AgentsAPI
	}
)

func NewAgentSendMessage(conversationRepository domain.ConversationRepository, agentsAPI agents.AgentsAPI) AgentSendMessage {
	return &agentSendMessage{conversationRepository, agentsAPI}
}

func (uc *agentSendMessage) Execute(ctx context.Context, input AgentSendMessageInput) error {
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

	err = conversation.AgentSendMessage(
		input.MessageID,
		agentID,
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
