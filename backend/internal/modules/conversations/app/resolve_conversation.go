package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
)

var ErrResolvingConversation = errors.New("resolving conversation failed")

type ResolveConversationInput struct {
	ConversationID uuid.UUID
	AgentID        uuid.UUID
	ResolvedAt     time.Time
}

type ResolveConversation interface {
	Execute(ctx context.Context, input ResolveConversationInput) error
}

type resolveConversation struct {
	conversationRepository domain.ConversationRepository
	usersAPI               users.UsersAPI
}

func NewResolveConversation(conversationRepository domain.ConversationRepository, usersAPI users.UsersAPI) ResolveConversation {
	return &resolveConversation{conversationRepository: conversationRepository, usersAPI: usersAPI}
}

func (uc *resolveConversation) Execute(ctx context.Context, input ResolveConversationInput) error {
	agentID, err := uc.usersAPI.FindUserByID(ctx, input.AgentID)
	if err != nil {
		return uc.joinErr(err)
	}

	conversation, err := uc.conversationRepository.FindWithoutMessages(ctx, input.ConversationID)
	if err != nil {
		return uc.joinErr(err)
	}

	if conversation == nil {
		return uc.joinErr(domain.ErrConversationNotFound)
	}

	if err := conversation.ResolveConversation(agentID, input.ResolvedAt); err != nil {
		if errors.Is(err, domain.ErrConversationFinished) {
			return nil
		}

		return uc.joinErr(err)
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	return nil
}

func (uc *resolveConversation) joinErr(errs ...error) error {
	return errors.Join(ErrResolvingConversation, errors.Join(errs...))
}
