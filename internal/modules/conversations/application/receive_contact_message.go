package application

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	ReceiveContactMessageInput struct {
		ConversationID    uuid.UUID
		MessageID         uuid.UUID
		ExternalMessageID string
		ExternalContactID string
		Text              string
		ReceivedAt        time.Time
	}

	ReceiveContactMessage interface {
		Execute(ctx context.Context, input ReceiveContactMessageInput) error
	}

	receiveContactMessage struct {
		conversationRepository domain.ConversationRepository
		contactsAPI            contacts.ContactsAPI
	}
)

func NewReceiveContactMessage(conversationRepository domain.ConversationRepository, contactsAPI contacts.ContactsAPI) ReceiveContactMessage {
	return &receiveContactMessage{conversationRepository, contactsAPI}
}

func (uc *receiveContactMessage) Execute(ctx context.Context, input ReceiveContactMessageInput) error {
	contactID, err := uc.contactsAPI.GetOrCreateContactIDByExternalID(ctx, input.ExternalContactID)
	if err != nil {
		return uc.wrapError(err)
	}

	conversation, err := uc.conversationRepository.FindLastOpenByContactID(ctx, contactID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		conversation, err = domain.NewConversation(input.ConversationID, domain.Pending, make(map[uuid.UUID]*domain.Message), 0, nil, contactID, input.ReceivedAt, nil, nil)
		if err != nil {
			return uc.wrapError(err)
		}
	}

	_, err = conversation.ReceiveContactMessage(input.MessageID, contactID, input.ExternalMessageID, input.Text, input.ReceivedAt)
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
