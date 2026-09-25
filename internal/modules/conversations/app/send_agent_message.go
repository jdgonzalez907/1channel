package app

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/agents"
	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

var ErrSendingAgentMessage = errors.New("sending agent message failed")

type SendAgentMessageInput struct {
	ConversationID uuid.UUID
	MessageID      uuid.UUID
	AgentID        uuid.UUID
	Text           string
	SentAt         time.Time
}

type SendAgentMessage interface {
	Execute(ctx context.Context, input SendAgentMessageInput) error
}

type sendAgentMessage struct {
	conversationRepository domain.ConversationRepository
	agentsAPI              agents.AgentsAPI
	contactsAPI            contacts.ContactsAPI
	messageSender          domain.AgentMessageSender
}

func NewSendAgentMessage(
	conversationRepository domain.ConversationRepository,
	agentsAPI agents.AgentsAPI,
	contactsAPI contacts.ContactsAPI,
	messageSender domain.AgentMessageSender,
) SendAgentMessage {
	return &sendAgentMessage{
		conversationRepository: conversationRepository,
		agentsAPI:              agentsAPI,
		contactsAPI:            contactsAPI,
		messageSender:          messageSender,
	}
}

func (uc *sendAgentMessage) Execute(ctx context.Context, input SendAgentMessageInput) error {
	agentID, err := uc.agentsAPI.FindAgentByID(ctx, input.AgentID)
	if err != nil {
		return uc.joinErr(err)
	}

	conversation, err := uc.conversationRepository.FindWithoutMessages(ctx, input.ConversationID)
	if err != nil {
		return uc.joinErr(err)
	}

	if conversation == nil {
		return uc.joinErr(domain.ErrConversationNotFound)
	}

	text := input.Text
	message, err := domain.NewMessage(
		input.MessageID,
		domain.MessageStatusSent,
		domain.MessageTypeText,
		&text,
		&agentID,
		nil,
		nil,
		input.SentAt,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return uc.joinErr(err)
	}

	if err := conversation.SendAgentMessage(agentID, message, input.SentAt); err != nil {
		return uc.joinErr(err)
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	contactID := conversation.ContactID()
	if contactID == nil {
		return uc.joinErr(domain.ErrConversationHasNoContact)
	}

	externalContactID, err := uc.contactsAPI.FindExternalContactIDByContactID(ctx, *contactID)
	if err != nil {
		return uc.joinErr(err)
	}

	externalMessageID, err := uc.messageSender.Send(ctx, externalContactID, message)
	if err != nil {
		if markErr := conversation.MarkAgentMessageFailed(agentID, message.ID(), input.SentAt); markErr != nil {
			return uc.joinErr(err, markErr)
		}

		if saveErr := uc.conversationRepository.Save(ctx, conversation); saveErr != nil {
			return uc.joinErr(err, saveErr)
		}

		return uc.joinErr(err)
	}

	if err := conversation.AssignAgentMessageExternalID(message.ID(), externalMessageID, input.SentAt); err != nil {
		return uc.joinErr(err)
	}

	if err := uc.conversationRepository.Save(ctx, conversation); err != nil {
		return uc.joinErr(err)
	}

	return nil
}

func (uc *sendAgentMessage) joinErr(errs ...error) error {
	return errors.Join(ErrSendingAgentMessage, errors.Join(errs...))
}
