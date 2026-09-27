package app

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
)

func TestSendAgentMessage_Execute(t *testing.T) {
	convID := uuid.NewV7()
	messageID := uuid.NewV7()
	contactID := uuid.NewV7()
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()
	externalContactID := "5491112345678"
	externalMessageID := "wamid.agent.001"
	now := time.Now()
	finishedAt := now.Add(-time.Hour)
	repoErr := errors.New("repository failure")
	agentsErr := errors.New("agents api failure")
	contactsErr := errors.New("contacts api failure")
	sendErr := errors.New("send failure")

	tests := []struct {
		name    string
		text    *string
		setup   func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation
		wantErr error
		check   func(t *testing.T, conv *domain.Conversation)
	}{
		{
			name: "sends message successfully",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Twice()
				contactsAPI.On("FindExternalContactIDByContactID", mock.Anything, contactID).Return(externalContactID, nil).Once()
				sender.On("Send", mock.Anything, externalContactID, mock.Anything).Return(externalMessageID, nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				messages := conv.Messages()
				assert.Len(t, messages, 1)
				assert.NotNil(t, messages[0].ExternalID())
				assert.Equal(t, externalMessageID, *messages[0].ExternalID())
			},
		},
		{
			name: "claims pending conversation",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Twice()
				contactsAPI.On("FindExternalContactIDByContactID", mock.Anything, contactID).Return(externalContactID, nil).Once()
				sender.On("Send", mock.Anything, externalContactID, mock.Anything).Return(externalMessageID, nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.ConversationStatusAssigned, conv.Status())
				assert.Equal(t, agentID, *conv.AgentID())
			},
		},
		{
			name: "agent not owner",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &otherAgentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationAgentNotOwner,
		},
		{
			name: "conversation not accepting messages",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusExpired, &agentID, &contactID, &finishedAt)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationNotAcceptingMessages,
		},
		{
			name: "send failure marks message failed",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Twice()
				contactsAPI.On("FindExternalContactIDByContactID", mock.Anything, contactID).Return(externalContactID, nil).Once()
				sender.On("Send", mock.Anything, externalContactID, mock.Anything).Return("", sendErr).Once()
				return conv
			},
			wantErr: sendErr,
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.MessageStatusFailed, conv.Messages()[0].Status())
			},
		},
		{
			name: "conversation not found",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(nil, nil).Once()
				return nil
			},
			wantErr: domain.ErrConversationNotFound,
		},
		{
			name: "agents api error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(uuid.Nil(), agentsErr).Once()
				return nil
			},
			wantErr: agentsErr,
		},
		{
			name: "contacts api error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				contactsAPI.On("FindExternalContactIDByContactID", mock.Anything, contactID).Return("", contactsErr).Once()
				return conv
			},
			wantErr: contactsErr,
		},
		{
			name: "find error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(nil, repoErr).Once()
				return nil
			},
			wantErr: repoErr,
		},
		{
			name: "invalid message text",
			text: strPtr(""),
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrMessageEmptyText,
		},
		{
			name: "conversation without contact",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, nil, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationHasNoContact,
		},
		{
			name: "send failure and failed save after marking failed",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				repo.On("Save", mock.Anything, conv).Return(repoErr).Once()
				contactsAPI.On("FindExternalContactIDByContactID", mock.Anything, contactID).Return(externalContactID, nil).Once()
				sender.On("Send", mock.Anything, externalContactID, mock.Anything).Return("", sendErr).Once()
				return conv
			},
			wantErr: sendErr,
		},
		{
			name: "save error after assigning external id",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				repo.On("Save", mock.Anything, conv).Return(repoErr).Once()
				contactsAPI.On("FindExternalContactIDByContactID", mock.Anything, contactID).Return(externalContactID, nil).Once()
				sender.On("Send", mock.Anything, externalContactID, mock.Anything).Return(externalMessageID, nil).Once()
				return conv
			},
			wantErr: repoErr,
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, externalMessageID, *conv.Messages()[0].ExternalID())
			},
		},
		{
			name: "sender returns empty external id",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				contactsAPI.On("FindExternalContactIDByContactID", mock.Anything, contactID).Return(externalContactID, nil).Once()
				sender.On("Send", mock.Anything, externalContactID, mock.Anything).Return("", nil).Once()
				return conv
			},
			wantErr: domain.ErrMessageExternalIDInvalid,
		},
		{
			name: "save error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, agentsAPI *users.MockUsersAPI, contactsAPI *contacts.MockContactsAPI, sender *domain.MockAgentMessageSender) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				agentsAPI.On("FindUserByID", mock.Anything, agentID).Return(agentID, nil).Once()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(repoErr).Once()
				return conv
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &domain.MockConversationRepository{}
			agentsAPI := &users.MockUsersAPI{}
			contactsAPI := &contacts.MockContactsAPI{}
			sender := &domain.MockAgentMessageSender{}
			conv := tt.setup(t, repo, agentsAPI, contactsAPI, sender)
			uc := NewSendAgentMessage(repo, agentsAPI, contactsAPI, sender)

			text := "hello"
			if tt.text != nil {
				text = *tt.text
			}

			// Act
			err := uc.Execute(context.Background(), SendAgentMessageInput{
				ConversationID: convID,
				MessageID:      messageID,
				AgentID:        agentID,
				Text:           text,
				SentAt:         now,
			})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrSendingAgentMessage)
			} else {
				assert.NoError(t, err)
			}
			if tt.check != nil {
				tt.check(t, conv)
			}
			repo.AssertExpectations(t)
			agentsAPI.AssertExpectations(t)
			contactsAPI.AssertExpectations(t)
			sender.AssertExpectations(t)
		})
	}
}
