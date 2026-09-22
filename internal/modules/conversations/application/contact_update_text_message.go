package application

import (
	"context"
	"errors"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	ContactUpdateTextMessageInput struct {
		ExternalMessageID string
		ExternalContactID string
		Text              string
		UpdatedAt         time.Time
	}

	ContactUpdateTextMessage interface {
		Execute(ctx context.Context, input ContactUpdateTextMessageInput) error
	}

	contactUpdateTextMessage struct {
		conversationRepository domain.ConversationRepository
		contactsAPI            contacts.ContactsAPI
	}
)

func NewContactUpdateTextMessage(conversationRepository domain.ConversationRepository, contactsAPI contacts.ContactsAPI) ContactUpdateTextMessage {
	return &contactUpdateTextMessage{conversationRepository, contactsAPI}
}

func (uc *contactUpdateTextMessage) Execute(ctx context.Context, input ContactUpdateTextMessageInput) error {
	contactID, err := uc.contactsAPI.GetOrCreateContactIDByExternalID(ctx, input.ExternalContactID)
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

	err = conversation.ContactUpdateTextMessage(contactID, input.ExternalMessageID, input.Text, input.UpdatedAt)
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
