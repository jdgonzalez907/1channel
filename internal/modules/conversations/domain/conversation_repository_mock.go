package domain

import (
	"context"

	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockConversationRepository struct {
	mock.Mock
}

func (m *MockConversationRepository) FindWithoutMessages(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	args := m.Called(ctx, id)

	conversation, _ := args.Get(0).(*Conversation)
	return conversation, args.Error(1)
}

func (m *MockConversationRepository) FindWithContactUnreadMessagesByID(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	args := m.Called(ctx, id)

	conversation, _ := args.Get(0).(*Conversation)
	return conversation, args.Error(1)
}

func (m *MockConversationRepository) FindWithMessageByExternalID(ctx context.Context, externalMessageID string) (*Conversation, error) {
	args := m.Called(ctx, externalMessageID)

	conversation, _ := args.Get(0).(*Conversation)
	return conversation, args.Error(1)
}

func (m *MockConversationRepository) FindOpenByContactID(ctx context.Context, contactID uuid.UUID) (*Conversation, error) {
	args := m.Called(ctx, contactID)

	conversation, _ := args.Get(0).(*Conversation)
	return conversation, args.Error(1)
}

func (m *MockConversationRepository) Save(ctx context.Context, conversation *Conversation) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}
