package application

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	ReceiveContactMessageInput struct {
		ConversationID    uuid.UUID
		MessageID         uuid.UUID
		ExternalMessageID string
		ContactID         uuid.UUID
		Text              string
		ReceivedAt        time.Time
	}

	ReceiveContactMessage interface {
		Execute(ctx context.Context, input ReceiveContactMessageInput) error
	}

	receiveContactMessage struct {
		conversationRepository domain.ConversationRepository
	}
)

func NewReceiveContactMessage(conversationRepository domain.ConversationRepository) ReceiveContactMessage {
	return &receiveContactMessage{conversationRepository}
}

func (uc *receiveContactMessage) Execute(ctx context.Context, input ReceiveContactMessageInput) error {
	conversation, err := uc.conversationRepository.FindLastOpenByContactID(ctx, input.ContactID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		conversation, err = domain.NewConversation(input.ConversationID, domain.Pending, make(map[uuid.UUID]*domain.Message), 0, nil, input.ContactID, input.ReceivedAt, nil, nil)
		if err != nil {
			return uc.wrapError(err)
		}
	}

	err = conversation.ReceiveContactMessage(input.MessageID, input.ContactID, input.ExternalMessageID, input.Text, input.ReceivedAt)
	if err != nil {
		if errors.Is(err, domain.ErrMessageAlreadyExists) {
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

func (uc *receiveContactMessage) wrapError(err error) error {
	return errors.Join(errors.New("error receiving contact message"), err)
}
