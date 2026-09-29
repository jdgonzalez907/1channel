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
)

func TestReceiveContactMessageRead_Execute(t *testing.T) {
	convID := uuid.NewV7()
	contactID := uuid.NewV7()
	agentID := uuid.NewV7()
	externalContactID := "5491112345678"
	externalMessageID := "wamid.agent.001"
	now := time.Now()
	repoErr := errors.New("repository failure")
	contactsErr := errors.New("contacts api failure")

	tests := []struct {
		name       string
		setup      func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation
		wantErr    error
		wantNoSave bool
		check      func(t *testing.T, conv *domain.Conversation)
	}{
		{
			name: "applies read",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				message := buildAgentMessageWithExternalID(t, uuid.NewV7(), agentID, externalMessageID)
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, message)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.MessageStatusRead, conv.Messages()[0].Status())
				assert.NotNil(t, conv.Messages()[0].ReadAt())
			},
		},
		{
			name: "external id not found",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				message := buildAgentMessageWithExternalID(t, uuid.NewV7(), agentID, "wamid.other")
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil, message)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID, mock.Anything).Return(contactID, nil).Once()
				repo.On("FindWithMessageByExternalID", mock.Anything, externalMessageID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrMessageNotFound,
		},
		{
			name: "contact message cannot be read by contact",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				message := buildContactMessage(t, uuid.NewV7(), contactID, strPtr(externalMessageID), nil)
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
				message := buildAgentMessageWithExternalID(t, uuid.NewV7(), agentID, externalMessageID)
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
			uc := NewReceiveContactMessageRead(repo, contactsAPI)

			// Act
			err := uc.Execute(context.Background(), ReceiveContactMessageReadInput{
				ExternalMessageID: externalMessageID,
				ExternalContactID: externalContactID,
				ReadAt:            now,
			})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrReceivingContactMessageRead)
			} else {
				assert.NoError(t, err)
			}
			if tt.check != nil {
				tt.check(t, conv)
			}
			if tt.wantNoSave {
				repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
			}
			repo.AssertExpectations(t)
			contactsAPI.AssertExpectations(t)
		})
	}
}
