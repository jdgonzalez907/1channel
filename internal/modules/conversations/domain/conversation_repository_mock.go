package domain

import (
	"context"
	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockConversationRepository struct {
	ConversationRepository
	mock.Mock
}

func (m *MockConversationRepository) FindByID(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	args := m.Called(ctx, id)
	return conversationFrom(args), args.Error(1)
}

func (m *MockConversationRepository) FindLastOpenByContactID(ctx context.Context, contactID uuid.UUID) (*Conversation, error) {
	args := m.Called(ctx, contactID)
	return conversationFrom(args), args.Error(1)
}

func (m *MockConversationRepository) FindWithSpecificMessageByExternalID(ctx context.Context, externalMessageID string) (*Conversation, error) {
	args := m.Called(ctx, externalMessageID)
	return conversationFrom(args), args.Error(1)
}

func (m *MockConversationRepository) FindWithSpecificMessageByMessageID(ctx context.Context, messageID uuid.UUID) (*Conversation, error) {
	args := m.Called(ctx, messageID)
	return conversationFrom(args), args.Error(1)
}

func (m *MockConversationRepository) Save(ctx context.Context, conversation *Conversation) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}

func conversationFrom(args mock.Arguments) *Conversation {
	value := args.Get(0)
	if value == nil {
		return nil
	}

	return value.(*Conversation)
}
