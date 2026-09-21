package conversations

import (
	"context"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/application"
)

type ConversationsAPI interface {
	ReceiveContactMessage(ctx context.Context, input ReceiveContactMessageInput) error
}

type ReceiveContactMessageInput struct {
	ExternalMessageID string
	ExternalContactID string
	Text              string
	ReceivedAt        time.Time
}

type conversationsAPI struct {
	receiveContactMessage application.ReceiveContactMessage
}

func NewConversationsAPI(receiveContactMessage application.ReceiveContactMessage) ConversationsAPI {
	return &conversationsAPI{receiveContactMessage}
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
