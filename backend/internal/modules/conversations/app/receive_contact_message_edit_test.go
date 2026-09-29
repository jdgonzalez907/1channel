package app

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

func TestReceiveContactMessageEdit_Execute(t *testing.T) {
	convID := uuid.NewV7()
	contactID := uuid.NewV7()
	agentID := uuid.NewV7()
	externalContactID := "5491112345678"
	externalMessageID := "wamid.contact.001"
	now := time.Now()
	repoErr := errors.New("repository failure")
	contactsErr := errors.New("contacts api failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation
		wantErr error
		check   func(t *testing.T, conv *domain.Conversation)
	}{
		{
			name: "applies edit",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				message := buildContactMessage(t, uuid.NewV7(), contactID, strPtr(externalMessageID), nil)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, message)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, "updated", *conv.Messages()[0].Text())
				assert.NotNil(t, conv.Messages()[0].EditedAt())
			},
		},
		{
			name: "stale edit is discarded",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				text := "hello"
				editedAt := now.Add(time.Hour)
				message, err := domain.NewMessage(uuid.NewV7(), domain.MessageStatusSent, domain.MessageTypeText, &text, nil, &contactID, strPtr(externalMessageID), now, nil, &editedAt, nil)
				require.NoError(t, err)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, message)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, "hello", *conv.Messages()[0].Text())
			},
		},
		{
			name: "external id not found",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				message := buildContactMessage(t, uuid.NewV7(), contactID, strPtr("wamid.other"), nil)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, message)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrMessageNotFound,
		},
		{
			name: "agent message cannot be edited by contact",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				message := buildAgentMessageWithExternalID(t, uuid.NewV7(), agentID, externalMessageID)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, message)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationContactNotOwner,
		},
		{
			name: "conversation not found",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(nil, nil).Once()
				return nil
			},
			wantErr: domain.ErrConversationNotFound,
		},
		{
			name: "contacts api error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(uuid.Nil(), contactsErr).Once()
				return nil
			},
			wantErr: contactsErr,
		},
		{
			name: "find error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(nil, repoErr).Once()
				return nil
			},
			wantErr: repoErr,
		},
		{
			name: "save error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				message := buildContactMessage(t, uuid.NewV7(), contactID, strPtr(externalMessageID), nil)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, message)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(conv, nil).Once()
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
			contactsAPI := &contacts.MockContactsAPI{}
			conv := tt.setup(t, repo, contactsAPI)
			uc := NewReceiveContactMessageEdit(repo, contactsAPI)

			// Act
			err := uc.Execute(context.Background(), ReceiveContactMessageEditInput{
				ExternalMessageID: externalMessageID,
				ExternalContactID: externalContactID,
				NewText:           "updated",
				EditedAt:          now,
			})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrReceivingContactMessageEdit)
			} else {
				assert.NoError(t, err)
			}
			if tt.check != nil {
				tt.check(t, conv)
			}
			repo.AssertExpectations(t)
			contactsAPI.AssertExpectations(t)
		})
	}
}
