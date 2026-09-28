package app

import (
	"context"
	"errors"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

var ErrReceivingContactMessageRead = errors.New("receiving contact message read failed")

type ReceiveContactMessageReadInput struct {
	ExternalMessageID string
	ExternalContactID string
	ReadAt            time.Time
}

type ReceiveContactMessageRead interface {
	Execute(ctx context.Context, input ReceiveContactMessageReadInput) error
}

type receiveContactMessageRead struct {
	conversationRepository domain.ConversationRepository
	contactsAPI            contacts.ContactsAPI
}

func NewReceiveContactMessageRead(conversationRepository domain.ConversationRepository, contactsAPI contacts.ContactsAPI) ReceiveContactMessageRead {
	return &receiveContactMessageRead{conversationRepository: conversationRepository, contactsAPI: contactsAPI}
}

func (uc *receiveContactMessageRead) Execute(ctx context.Context, input ReceiveContactMessageReadInput) error {
	contactID, err := uc.contactsAPI.GetOrCreateContactIDByExternalID(ctx, input.ExternalContactID)
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

	if err := conversation.ReceiveContactMessageRead(contactID, input.ExternalMessageID, input.ReadAt); err != nil {
		return uc.joinErr(err)
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	return nil
}

func (uc *receiveContactMessageRead) joinErr(errs ...error) error {
	return errors.Join(ErrReceivingContactMessageRead, errors.Join(errs...))
}
