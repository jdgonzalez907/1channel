package application

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	ContactUpdateTextMessageInput struct {
		ExternalMessageID string
		ContactID         uuid.UUID
		Text              string
		UpdatedAt         time.Time
	}

	ContactUpdateTextMessage interface {
		Execute(ctx context.Context, input ContactUpdateTextMessageInput) error
	}

	contactUpdateTextMessage struct {
		conversationRepository domain.ConversationRepository
	}
)

func NewContactUpdateTextMessage(conversationRepository domain.ConversationRepository) ContactUpdateTextMessage {
	return &contactUpdateTextMessage{conversationRepository}
}

func (uc *contactUpdateTextMessage) Execute(ctx context.Context, input ContactUpdateTextMessageInput) error {
	conversation, err := uc.conversationRepository.FindWithSpecificMessageByExternalID(ctx, input.ExternalMessageID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.ContactUpdateTextMessage(input.ContactID, input.ExternalMessageID, input.Text, input.UpdatedAt)
	if err != nil {
		return uc.wrapError(err)
	}

	err = uc.conversationRepository.Save(ctx, conversation)
	if err != nil {
		return uc.wrapError(err)
	}

	return nil
}

func (uc *contactUpdateTextMessage) wrapError(err error) error {
	return errors.Join(errors.New("error contact updating text message"), err)
}
