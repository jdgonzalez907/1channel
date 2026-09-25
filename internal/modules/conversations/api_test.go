package conversations

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type apiMocks struct {
	receive       *app.MockReceiveContactMessage
	receiveEdit   *app.MockReceiveContactMessageEdit
	receiveDelete *app.MockReceiveContactMessageDelete
	send          *app.MockSendAgentMessage
	read          *app.MockAgentReadConversation
	expire        *app.MockExpireConversation
	resolve       *app.MockResolveConversation
}

func newAPIWithMocks() (ConversationsAPI, *apiMocks) {
	m := &apiMocks{
		receive:       &app.MockReceiveContactMessage{},
		receiveEdit:   &app.MockReceiveContactMessageEdit{},
		receiveDelete: &app.MockReceiveContactMessageDelete{},
		send:          &app.MockSendAgentMessage{},
		read:          &app.MockAgentReadConversation{},
		expire:        &app.MockExpireConversation{},
		resolve:       &app.MockResolveConversation{},
	}

	api := NewConversationsAPI(m.receive, m.receiveEdit, m.receiveDelete, m.send, m.read, m.expire, m.resolve)

	return api, m
}

func TestNewConversationsAPI(t *testing.T) {
	// Act
	api, _ := newAPIWithMocks()

	// Assert
	assert.NotNil(t, api)
}

func TestConversationsAPI_ReceiveContactMessage(t *testing.T) {
	externalMessageID := "wamid.1"
	externalContactID := "5491112345678"
	now := time.Now()
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockReceiveContactMessage)
		input   ReceiveContactMessageInput
		wantErr error
	}{
		{
			name: "success - generates ids and maps input",
			setup: func(t *testing.T, m *app.MockReceiveContactMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ReceiveContactMessageInput) bool {
					return in.ConversationID != uuid.Nil() &&
						in.MessageID != uuid.Nil() &&
						in.ExternalMessageID != nil && *in.ExternalMessageID == externalMessageID &&
						in.ExternalContactID == externalContactID &&
						in.ReceivedAt.Equal(now)
				})).Return(nil).Once()
			},
			input: ReceiveContactMessageInput{
				ExternalMessageID: &externalMessageID,
				ExternalContactID: externalContactID,
				Text:              "hello",
				ReceivedAt:        now,
			},
		},
		{
			name: "failure - propagates error",
			setup: func(t *testing.T, m *app.MockReceiveContactMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(ucErr).Once()
			},
			input:   ReceiveContactMessageInput{ExternalContactID: externalContactID, Text: "hello", ReceivedAt: now},
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api, m := newAPIWithMocks()
			tt.setup(t, m.receive)

			// Act
			err := api.ReceiveContactMessage(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			m.receive.AssertExpectations(t)
		})
	}
}

func TestConversationsAPI_SendAgentMessage(t *testing.T) {
	conversationID := uuid.NewV7()
	agentID := uuid.NewV7()
	now := time.Now()
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockSendAgentMessage)
		wantErr error
	}{
		{
			name: "success - generates message id",
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(in app.SendAgentMessageInput) bool {
					return in.ConversationID == conversationID &&
						in.MessageID != uuid.Nil() &&
						in.AgentID == agentID &&
						in.Text == "hello"
				})).Return(nil).Once()
			},
		},
		{
			name: "failure - propagates error",
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(ucErr).Once()
			},
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api, m := newAPIWithMocks()
			tt.setup(t, m.send)

			// Act
			err := api.SendAgentMessage(context.Background(), SendAgentMessageInput{
				ConversationID: conversationID,
				AgentID:        agentID,
				Text:           "hello",
				SentAt:         now,
			})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			m.send.AssertExpectations(t)
		})
	}
}

func TestConversationsAPI_ReceiveContactMessageEdit(t *testing.T) {
	now := time.Now()
	ucErr := errors.New("use case failure")
	input := ReceiveContactMessageEditInput{ExternalMessageID: "wamid.1", ExternalContactID: "54911", NewText: "hi", EditedAt: now}

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockReceiveContactMessageEdit)
		wantErr error
	}{
		{
			name: "success - delegates",
			setup: func(t *testing.T, m *app.MockReceiveContactMessageEdit) {
				t.Helper()
				m.On("Execute", mock.Anything, app.ReceiveContactMessageEditInput{
					ExternalMessageID: input.ExternalMessageID,
					ExternalContactID: input.ExternalContactID,
					NewText:           input.NewText,
					EditedAt:          input.EditedAt,
				}).Return(nil).Once()
			},
		},
		{
			name: "failure - propagates error",
			setup: func(t *testing.T, m *app.MockReceiveContactMessageEdit) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(ucErr).Once()
			},
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api, m := newAPIWithMocks()
			tt.setup(t, m.receiveEdit)

			// Act
			err := api.ReceiveContactMessageEdit(context.Background(), input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			m.receiveEdit.AssertExpectations(t)
		})
	}
}

func TestConversationsAPI_ReceiveContactMessageDelete(t *testing.T) {
	now := time.Now()
	ucErr := errors.New("use case failure")
	input := ReceiveContactMessageDeleteInput{ExternalMessageID: "wamid.1", ExternalContactID: "54911", DeletedAt: now}

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockReceiveContactMessageDelete)
		wantErr error
	}{
		{
			name: "success - delegates",
			setup: func(t *testing.T, m *app.MockReceiveContactMessageDelete) {
				t.Helper()
				m.On("Execute", mock.Anything, app.ReceiveContactMessageDeleteInput{
					ExternalMessageID: input.ExternalMessageID,
					ExternalContactID: input.ExternalContactID,
					DeletedAt:         input.DeletedAt,
				}).Return(nil).Once()
			},
		},
		{
			name: "failure - propagates error",
			setup: func(t *testing.T, m *app.MockReceiveContactMessageDelete) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(ucErr).Once()
			},
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api, m := newAPIWithMocks()
			tt.setup(t, m.receiveDelete)

			// Act
			err := api.ReceiveContactMessageDelete(context.Background(), input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			m.receiveDelete.AssertExpectations(t)
		})
	}
}

func TestConversationsAPI_AgentReadConversation(t *testing.T) {
	conversationID := uuid.NewV7()
	agentID := uuid.NewV7()
	now := time.Now()
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockAgentReadConversation)
		wantErr error
	}{
		{
			name: "success - delegates",
			setup: func(t *testing.T, m *app.MockAgentReadConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, app.AgentReadConversationInput{
					ConversationID: conversationID,
					AgentID:        agentID,
					ReadAt:         now,
				}).Return(nil).Once()
			},
		},
		{
			name: "failure - propagates error",
			setup: func(t *testing.T, m *app.MockAgentReadConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(ucErr).Once()
			},
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api, m := newAPIWithMocks()
			tt.setup(t, m.read)

			// Act
			err := api.AgentReadConversation(context.Background(), AgentReadConversationInput{ConversationID: conversationID, AgentID: agentID, ReadAt: now})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			m.read.AssertExpectations(t)
		})
	}
}

func TestConversationsAPI_ExpireConversation(t *testing.T) {
	conversationID := uuid.NewV7()
	now := time.Now()
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockExpireConversation)
		wantErr error
	}{
		{
			name: "success - delegates",
			setup: func(t *testing.T, m *app.MockExpireConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, app.ExpireConversationInput{ConversationID: conversationID, ExpiredAt: now}).Return(nil).Once()
			},
		},
		{
			name: "failure - propagates error",
			setup: func(t *testing.T, m *app.MockExpireConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(ucErr).Once()
			},
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api, m := newAPIWithMocks()
			tt.setup(t, m.expire)

			// Act
			err := api.ExpireConversation(context.Background(), ExpireConversationInput{ConversationID: conversationID, ExpiredAt: now})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			m.expire.AssertExpectations(t)
		})
	}
}

func TestConversationsAPI_ResolveConversation(t *testing.T) {
	conversationID := uuid.NewV7()
	agentID := uuid.NewV7()
	now := time.Now()
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockResolveConversation)
		wantErr error
	}{
		{
			name: "success - delegates",
			setup: func(t *testing.T, m *app.MockResolveConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, app.ResolveConversationInput{ConversationID: conversationID, AgentID: agentID, ResolvedAt: now}).Return(nil).Once()
			},
		},
		{
			name: "failure - propagates error",
			setup: func(t *testing.T, m *app.MockResolveConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(ucErr).Once()
			},
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			api, m := newAPIWithMocks()
			tt.setup(t, m.resolve)

			// Act
			err := api.ResolveConversation(context.Background(), ResolveConversationInput{ConversationID: conversationID, AgentID: agentID, ResolvedAt: now})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			m.resolve.AssertExpectations(t)
		})
	}
}
