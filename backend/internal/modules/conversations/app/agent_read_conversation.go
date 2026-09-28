package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
)

var ErrAgentReadingConversation = errors.New("agent reading conversation failed")

type AgentReadConversationInput struct {
	ConversationID uuid.UUID
	AgentID        uuid.UUID
	ReadAt         time.Time
}

type AgentReadConversation interface {
	Execute(ctx context.Context, input AgentReadConversationInput) error
}

type agentReadConversation struct {
	conversationRepository domain.ConversationRepository
	usersAPI               users.UsersAPI
}

func NewAgentReadConversation(conversationRepository domain.ConversationRepository, usersAPI users.UsersAPI) AgentReadConversation {
	return &agentReadConversation{conversationRepository: conversationRepository, usersAPI: usersAPI}
}

func (uc *agentReadConversation) Execute(ctx context.Context, input AgentReadConversationInput) error {
	agentID, err := uc.usersAPI.FindUserByID(ctx, input.AgentID)
	if err != nil {
		return uc.joinErr(err)
	}

	conversation, err := uc.conversationRepository.FindWithContactUnreadMessagesByID(ctx, input.ConversationID)
	if err != nil {
		return uc.joinErr(err)
	}

	if conversation == nil {
		return uc.joinErr(domain.ErrConversationNotFound)
	}

	if err := conversation.AgentReadConversation(agentID, input.ReadAt); err != nil {
		return uc.joinErr(err)
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	return nil
}

func (uc *agentReadConversation) joinErr(errs ...error) error {
	return errors.Join(ErrAgentReadingConversation, errors.Join(errs...))
}
