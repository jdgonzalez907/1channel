package conversations

import (
	"context"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/application"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type ConversationsAPI interface {
	ReceiveContactMessage(ctx context.Context, input ReceiveContactMessageInput) error
	UpdateAgentMessageStatus(ctx context.Context, input UpdateAgentMessageStatusInput) error
}

type ReceiveContactMessageInput struct {
	ExternalMessageID string
	ExternalContactID string
	Text              string
	ReceivedAt        time.Time
}

type UpdateAgentMessageStatusInput struct {
	MessageID uuid.UUID
	Status    string
	Timestamp time.Time
}

type conversationsAPI struct {
	receiveContactMessage    application.ReceiveContactMessage
	updateAgentMessageStatus application.UpdateAgentMessageStatus
}

func NewConversationsAPI(
	receiveContactMessage application.ReceiveContactMessage,
	updateAgentMessageStatus application.UpdateAgentMessageStatus,
) ConversationsAPI {
	return &conversationsAPI{receiveContactMessage, updateAgentMessageStatus}
}

func (a *conversationsAPI) ReceiveContactMessage(ctx context.Context, input ReceiveContactMessageInput) error {
	return a.receiveContactMessage.Execute(ctx, application.ReceiveContactMessageInput{
		ConversationID:    uuid.NewV7(),
		MessageID:         uuid.NewV7(),
		ExternalMessageID: input.ExternalMessageID,
		ExternalContactID: input.ExternalContactID,
		Text:              input.Text,
		ReceivedAt:        input.ReceivedAt,
	})
}

func (a *conversationsAPI) UpdateAgentMessageStatus(ctx context.Context, input UpdateAgentMessageStatusInput) error {
	status, err := domain.NewMessageStatus(input.Status)
	if err != nil {
		return err
	}

	return a.updateAgentMessageStatus.Execute(ctx, application.UpdateAgentMessageStatusInput{
		MessageID: input.MessageID,
		Status:    status,
		Timestamp: input.Timestamp,
	})
}
