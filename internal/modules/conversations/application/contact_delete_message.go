package application

import (
	"context"
	"errors"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	ContactDeleteMessageInput struct {
		ExternalMessageID string
		ExternalContactID string
		DeletedAt         time.Time
	}

	ContactDeleteMessage interface {
		Execute(ctx context.Context, input ContactDeleteMessageInput) error
	}

	contactDeleteMessage struct {
		conversationRepository domain.ConversationRepository
		contactsAPI            contacts.ContactsAPI
	}
)

func NewContactDeleteMessage(conversationRepository domain.ConversationRepository, contactsAPI contacts.ContactsAPI) ContactDeleteMessage {
	return &contactDeleteMessage{conversationRepository, contactsAPI}
}

func (uc *contactDeleteMessage) Execute(ctx context.Context, input ContactDeleteMessageInput) error {
	contactID, err := uc.contactsAPI.GetOrCreateContactByExternalID(ctx, input.ExternalContactID)
	if err != nil {
		return uc.wrapError(err)
	}

	conversation, err := uc.conversationRepository.FindWithSpecificMessageByExternalID(ctx, input.ExternalMessageID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.ContactDeleteMessage(contactID, input.ExternalMessageID, input.DeletedAt)
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
