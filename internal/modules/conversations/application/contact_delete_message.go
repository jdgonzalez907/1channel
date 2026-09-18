package application

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	ContactDeleteMessageInput struct {
		ExternalMessageID string
		ContactID         uuid.UUID
		DeletedAt         time.Time
	}

	ContactDeleteMessage interface {
		Execute(ctx context.Context, input ContactDeleteMessageInput) error
	}

	contactDeleteMessage struct {
		conversationRepository domain.ConversationRepository
	}
)

func NewContactDeleteMessage(conversationRepository domain.ConversationRepository) ContactDeleteMessage {
	return &contactDeleteMessage{conversationRepository}
}

func (uc *contactDeleteMessage) Execute(ctx context.Context, input ContactDeleteMessageInput) error {
	conversation, err := uc.conversationRepository.FindWithSpecificMessageByExternalID(ctx, input.ExternalMessageID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.ContactDeleteMessage(input.ContactID, input.ExternalMessageID, input.DeletedAt)
	if err != nil {
		return uc.wrapError(err)
	}

	err = uc.conversationRepository.Save(ctx, conversation)
	if err != nil {
		return uc.wrapError(err)
	}

	return nil
}

func (uc *contactDeleteMessage) wrapError(err error) error {
	return errors.Join(errors.New("error contact deleting text message"), err)
}
