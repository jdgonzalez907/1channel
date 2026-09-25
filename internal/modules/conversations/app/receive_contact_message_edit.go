package app

import (
	"context"
	"errors"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

var ErrReceivingContactMessageEdit = errors.New("receiving contact message edit failed")

type ReceiveContactMessageEditInput struct {
	ExternalMessageID string
	ExternalContactID string
	NewText           string
	EditedAt          time.Time
}

type ReceiveContactMessageEdit interface {
	Execute(ctx context.Context, input ReceiveContactMessageEditInput) error
}

type receiveContactMessageEdit struct {
	conversationRepository domain.ConversationRepository
	contactsAPI            contacts.ContactsAPI
}

func NewReceiveContactMessageEdit(conversationRepository domain.ConversationRepository, contactsAPI contacts.ContactsAPI) ReceiveContactMessageEdit {
	return &receiveContactMessageEdit{conversationRepository: conversationRepository, contactsAPI: contactsAPI}
}

func (uc *receiveContactMessageEdit) Execute(ctx context.Context, input ReceiveContactMessageEditInput) error {
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

	if err := conversation.ReceiveContactMessageEdit(contactID, input.ExternalMessageID, input.NewText, input.EditedAt); err != nil {
		return uc.joinErr(err)
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	return nil
}

func (uc *receiveContactMessageEdit) joinErr(errs ...error) error {
	return errors.Join(ErrReceivingContactMessageEdit, errors.Join(errs...))
}
