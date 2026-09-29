package app

import (
	"context"
	"errors"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

var ErrReceivingContactMessageDelete = errors.New("receiving contact message delete failed")

type ReceiveContactMessageDeleteInput struct {
	ExternalMessageID string
	ExternalContactID string
	DeletedAt         time.Time
}

type ReceiveContactMessageDelete interface {
	Execute(ctx context.Context, input ReceiveContactMessageDeleteInput) error
}

type receiveContactMessageDelete struct {
	conversationRepository domain.ConversationRepository
	contactsAPI            contacts.ContactsAPI
}

func NewReceiveContactMessageDelete(conversationRepository domain.ConversationRepository, contactsAPI contacts.ContactsAPI) ReceiveContactMessageDelete {
	return &receiveContactMessageDelete{conversationRepository: conversationRepository, contactsAPI: contactsAPI}
}

func (uc *receiveContactMessageDelete) Execute(ctx context.Context, input ReceiveContactMessageDeleteInput) error {
	contactID, err := uc.contactsAPI.GetOrCreateContactIDByExternalID(ctx, input.ExternalContactID, nil)
	if err != nil {
		return uc.joinErr(err)
	}

	conversation, err := uc.conversationRepository.FindWithMessageByExternalID(ctx, input.ExternalMessageID)
	if err != nil {
		return uc.joinErr(err)
	}

	if conversation == nil {
		return uc.joinErr(domain.ErrConversationNotFound)
	}

	if err := conversation.ReceiveContactMessageDelete(contactID, input.ExternalMessageID, input.DeletedAt); err != nil {
		if errors.Is(err, domain.ErrMessageAlreadyDeleted) {
			return nil
		}

		return uc.joinErr(err)
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	return nil
}

func (uc *receiveContactMessageDelete) joinErr(errs ...error) error {
	return errors.Join(ErrReceivingContactMessageDelete, errors.Join(errs...))
}
