package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

var ErrReceivingContactMessage = errors.New("receiving contact message failed")

type ReceiveContactMessageInput struct {
	ConversationID    uuid.UUID
	MessageID         uuid.UUID
	ExternalMessageID *string
	ExternalContactID string
	DisplayName       *string
	Text              string
	ReceivedAt        time.Time
}

type ReceiveContactMessage interface {
	Execute(ctx context.Context, input ReceiveContactMessageInput) error
}

type receiveContactMessage struct {
	conversationRepository domain.ConversationRepository
	contactsAPI            contacts.ContactsAPI
}

func NewReceiveContactMessage(conversationRepository domain.ConversationRepository, contactsAPI contacts.ContactsAPI) ReceiveContactMessage {
	return &receiveContactMessage{conversationRepository: conversationRepository, contactsAPI: contactsAPI}
}

func (uc *receiveContactMessage) Execute(ctx context.Context, input ReceiveContactMessageInput) error {
	contactID, err := uc.contactsAPI.GetOrCreateContactIDByExternalID(ctx, input.ExternalContactID, input.DisplayName)
	if err != nil {
		return uc.joinErr(err)
	}

	text := input.Text
	message, err := domain.NewMessage(
		input.MessageID,
		domain.MessageStatusSent,
		domain.MessageTypeText,
		&text,
		nil,
		&contactID,
		input.ExternalMessageID,
		input.ReceivedAt,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return uc.joinErr(err)
	}

	conversation, err := uc.conversationRepository.FindOpenWithMessageExternalIDsByContactID(ctx, contactID)
	if err != nil {
		return uc.joinErr(err)
	}

	if conversation == nil {
		conversation, err = domain.NewConversation(
			input.ConversationID,
			domain.ConversationStatusPending,
			[]*domain.Message{message},
			nil,
			&contactID,
			input.ReceivedAt,
			nil,
			nil,
		)
		if err != nil {
			return uc.joinErr(err)
		}
	} else {
		if err := conversation.ReceiveContactMessage(contactID, message, input.ReceivedAt); err != nil {
			if errors.Is(err, domain.ErrConversationDuplicateMessage) {
				return nil
			}

			return uc.joinErr(err)
		}
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	return nil
}

func (uc *receiveContactMessage) joinErr(errs ...error) error {
	return errors.Join(ErrReceivingContactMessage, errors.Join(errs...))
}
