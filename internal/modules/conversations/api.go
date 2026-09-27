package conversations

import (
	"context"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
)

type ConversationsAPI interface {
	ReceiveContactMessage(ctx context.Context, input ReceiveContactMessageInput) error
	ReceiveContactMessageEdit(ctx context.Context, input ReceiveContactMessageEditInput) error
	ReceiveContactMessageDelete(ctx context.Context, input ReceiveContactMessageDeleteInput) error
	SendAgentMessage(ctx context.Context, input SendAgentMessageInput) error
	AgentReadConversation(ctx context.Context, input AgentReadConversationInput) error
	ExpireConversation(ctx context.Context, input ExpireConversationInput) error
	ResolveConversation(ctx context.Context, input ResolveConversationInput) error
}

type ReceiveContactMessageInput struct {
	ExternalMessageID *string
	ExternalContactID string
	Text              string
	ReceivedAt        time.Time
}

type ReceiveContactMessageEditInput struct {
	ExternalMessageID string
	ExternalContactID string
	NewText           string
	EditedAt          time.Time
}

type ReceiveContactMessageDeleteInput struct {
	ExternalMessageID string
	ExternalContactID string
	DeletedAt         time.Time
}

type SendAgentMessageInput struct {
	ConversationID uuid.UUID
	AgentID        uuid.UUID
	Text           string
	SentAt         time.Time
}

type AgentReadConversationInput struct {
	ConversationID uuid.UUID
	AgentID        uuid.UUID
	ReadAt         time.Time
}

type ExpireConversationInput struct {
	ConversationID uuid.UUID
	ExpiredAt      time.Time
}

type ResolveConversationInput struct {
	ConversationID uuid.UUID
	AgentID        uuid.UUID
	ResolvedAt     time.Time
}

type conversationsAPI struct {
	receiveContactMessage       app.ReceiveContactMessage
	receiveContactMessageEdit   app.ReceiveContactMessageEdit
	receiveContactMessageDelete app.ReceiveContactMessageDelete
	sendAgentMessage            app.SendAgentMessage
	agentReadConversation       app.AgentReadConversation
	expireConversation          app.ExpireConversation
	resolveConversation         app.ResolveConversation
}

func NewConversationsAPI(
	receiveContactMessage app.ReceiveContactMessage,
	receiveContactMessageEdit app.ReceiveContactMessageEdit,
	receiveContactMessageDelete app.ReceiveContactMessageDelete,
	sendAgentMessage app.SendAgentMessage,
	agentReadConversation app.AgentReadConversation,
	expireConversation app.ExpireConversation,
	resolveConversation app.ResolveConversation,
) ConversationsAPI {
	return &conversationsAPI{
		receiveContactMessage:       receiveContactMessage,
		receiveContactMessageEdit:   receiveContactMessageEdit,
		receiveContactMessageDelete: receiveContactMessageDelete,
		sendAgentMessage:            sendAgentMessage,
		agentReadConversation:       agentReadConversation,
		expireConversation:          expireConversation,
		resolveConversation:         resolveConversation,
	}
}

func (a *conversationsAPI) ReceiveContactMessage(ctx context.Context, input ReceiveContactMessageInput) error {
	return a.receiveContactMessage.Execute(ctx, app.ReceiveContactMessageInput{
		ConversationID:    uuid.NewV7(),
		MessageID:         uuid.NewV7(),
		ExternalMessageID: input.ExternalMessageID,
		ExternalContactID: input.ExternalContactID,
		Text:              input.Text,
		ReceivedAt:        input.ReceivedAt,
	})
}

func (a *conversationsAPI) ReceiveContactMessageEdit(ctx context.Context, input ReceiveContactMessageEditInput) error {
	return a.receiveContactMessageEdit.Execute(ctx, app.ReceiveContactMessageEditInput{
		ExternalMessageID: input.ExternalMessageID,
		ExternalContactID: input.ExternalContactID,
		NewText:           input.NewText,
		EditedAt:          input.EditedAt,
	})
}

func (a *conversationsAPI) ReceiveContactMessageDelete(ctx context.Context, input ReceiveContactMessageDeleteInput) error {
	return a.receiveContactMessageDelete.Execute(ctx, app.ReceiveContactMessageDeleteInput{
		ExternalMessageID: input.ExternalMessageID,
		ExternalContactID: input.ExternalContactID,
		DeletedAt:         input.DeletedAt,
	})
}

func (a *conversationsAPI) SendAgentMessage(ctx context.Context, input SendAgentMessageInput) error {
	return a.sendAgentMessage.Execute(ctx, app.SendAgentMessageInput{
		ConversationID: input.ConversationID,
		MessageID:      uuid.NewV7(),
		AgentID:        input.AgentID,
		Text:           input.Text,
		SentAt:         input.SentAt,
	})
}

func (a *conversationsAPI) AgentReadConversation(ctx context.Context, input AgentReadConversationInput) error {
	return a.agentReadConversation.Execute(ctx, app.AgentReadConversationInput{
		ConversationID: input.ConversationID,
		AgentID:        input.AgentID,
		ReadAt:         input.ReadAt,
	})
}

func (a *conversationsAPI) ExpireConversation(ctx context.Context, input ExpireConversationInput) error {
	return a.expireConversation.Execute(ctx, app.ExpireConversationInput{
		ConversationID: input.ConversationID,
		ExpiredAt:      input.ExpiredAt,
	})
}

func (a *conversationsAPI) ResolveConversation(ctx context.Context, input ResolveConversationInput) error {
	return a.resolveConversation.Execute(ctx, app.ResolveConversationInput{
		ConversationID: input.ConversationID,
		AgentID:        input.AgentID,
		ResolvedAt:     input.ResolvedAt,
	})
}
