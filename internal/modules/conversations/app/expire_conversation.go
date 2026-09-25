package app

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

var ErrExpiringConversation = errors.New("expiring conversation failed")

type ExpireConversationInput struct {
	ConversationID uuid.UUID
	ExpiredAt      time.Time
}

type ExpireConversation interface {
	Execute(ctx context.Context, input ExpireConversationInput) error
}

type expireConversation struct {
	conversationRepository domain.ConversationRepository
}

func NewExpireConversation(conversationRepository domain.ConversationRepository) ExpireConversation {
	return &expireConversation{conversationRepository: conversationRepository}
}

func (uc *expireConversation) Execute(ctx context.Context, input ExpireConversationInput) error {
	conversation, err := uc.conversationRepository.FindWithoutMessages(ctx, input.ConversationID)
	if err != nil {
		return uc.joinErr(err)
	}

	if conversation == nil {
		return uc.joinErr(domain.ErrConversationNotFound)
	}

	if err := conversation.ExpireConversation(input.ExpiredAt); err != nil {
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

func (uc *expireConversation) joinErr(errs ...error) error {
	return errors.Join(ErrExpiringConversation, errors.Join(errs...))
}
