package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReceiveContactMessage_Execute(t *testing.T) {
	convID := uuid.NewV7()
	messageID := uuid.NewV7()
	contactID := uuid.NewV7()
	otherContactID := uuid.NewV7()
	externalContactID := "5491112345678"
	externalMessageID := "wamid.contact.001"
	now := time.Now()
	repoErr := errors.New("repository failure")
	contactsErr := errors.New("contacts api failure")

	tests := []struct {
		name                  string
		text                  *string
		invalidConversationID bool
		setup                 func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation
		wantErr               error
		check                 func(t *testing.T, conv *domain.Conversation)
	}{
		{
			name: "appends to active conversation",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &contactID, nil)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				repo.On("FindOpenWithMessageExternalIDsByContactID", mock.Anything, contactID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				messages := conv.Messages()
				assert.Len(t, messages, 1)
				assert.Equal(t, messageID, messages[0].ID())
				assert.NotNil(t, messages[0].ExternalID())
			},
		},
		{
			name: "creates new conversation when none active",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				repo.On("FindOpenWithMessageExternalIDsByContactID", mock.Anything, contactID).Return(nil, nil).Once()
				repo.On("Save", mock.Anything, mock.MatchedBy(func(c *domain.Conversation) bool {
					return c.Status() == domain.ConversationStatusPending &&
						c.ContactID() != nil && *c.ContactID() == contactID &&
						len(c.Messages()) == 1
				})).Return(nil).Once()
				return nil
			},
		},
		{
			name: "duplicate external id is idempotent",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				existing := buildContactMessage(t, uuid.NewV7(), contactID, strPtr(externalMessageID), nil)
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &contactID, nil, existing)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				repo.On("FindOpenWithMessageExternalIDsByContactID", mock.Anything, contactID).Return(conv, nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Len(t, conv.Messages(), 1)
			},
		},
		{
			name: "conversation of another contact",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &otherContactID, nil)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				repo.On("FindOpenWithMessageExternalIDsByContactID", mock.Anything, contactID).Return(conv, nil).Once()
				return conv
			},
			wantErr: domain.ErrConversationContactNotOwner,
		},
		{
			name: "invalid message text",
			text: strPtr(""),
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				return nil
			},
			wantErr: domain.ErrMessageEmptyText,
		},
		{
			name:                  "invalid conversation id",
			invalidConversationID: true,
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				repo.On("FindOpenWithMessageExternalIDsByContactID", mock.Anything, contactID).Return(nil, nil).Once()
				return nil
			},
			wantErr: domain.ErrConversationInvalidID,
		},
		{
			name: "contacts api error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(uuid.Nil(), contactsErr).Once()
				return nil
			},
			wantErr: contactsErr,
		},
		{
			name: "find error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				repo.On("FindOpenWithMessageExternalIDsByContactID", mock.Anything, contactID).Return(nil, repoErr).Once()
				return nil
			},
			wantErr: repoErr,
		},
		{
			name: "save error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository, contactsAPI *contacts.MockContactsAPI) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &contactID, nil)
				contactsAPI.On("GetOrCreateContactIDByExternalID", mock.Anything, externalContactID).Return(contactID, nil).Once()
				repo.On("FindOpenWithMessageExternalIDsByContactID", mock.Anything, contactID).Return(conv, nil).Once()
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
			uc := NewReceiveContactMessage(repo, contactsAPI)

			text := "hello"
			if tt.text != nil {
				text = *tt.text
			}

			conversationID := convID
			if tt.invalidConversationID {
				conversationID = uuid.Nil()
			}

			// Act
			err := uc.Execute(context.Background(), ReceiveContactMessageInput{
				ConversationID:    conversationID,
				MessageID:         messageID,
				ExternalMessageID: strPtr(externalMessageID),
				ExternalContactID: externalContactID,
				Text:              text,
				ReceivedAt:        now,
			})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrReceivingContactMessage)
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
