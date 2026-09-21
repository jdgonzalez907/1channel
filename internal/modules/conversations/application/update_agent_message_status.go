package application

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type (
	UpdateAgentMessageStatusInput struct {
		MessageID uuid.UUID
		Status    domain.MessageStatus
		Timestamp time.Time
	}

	UpdateAgentMessageStatus interface {
		Execute(ctx context.Context, input UpdateAgentMessageStatusInput) error
	}

	updateAgentMessageStatus struct {
		conversationRepository domain.ConversationRepository
	}
)

func NewUpdateAgentMessageStatus(conversationRepository domain.ConversationRepository) UpdateAgentMessageStatus {
	return &updateAgentMessageStatus{conversationRepository}
}

func (uc *updateAgentMessageStatus) Execute(ctx context.Context, input UpdateAgentMessageStatusInput) error {
	conversation, err := uc.conversationRepository.FindWithSpecificMessageByMessageID(ctx, input.MessageID)
	if err != nil {
		return uc.wrapError(err)
	}

	if conversation == nil {
		return uc.wrapError(domain.ErrConversationNotFound)
	}

	err = conversation.MarkAgentMessageStatus(input.MessageID, input.Status, input.Timestamp)
	if err != nil {
		return uc.wrapError(err)
	}

	err = uc.conversationRepository.Save(ctx, conversation)
	if err != nil {
		return uc.wrapError(err)
	}

	return nil
}

func (uc *updateAgentMessageStatus) wrapError(err error) error {
	return errors.Join(errors.New("error updating agent message status"), err)
}
