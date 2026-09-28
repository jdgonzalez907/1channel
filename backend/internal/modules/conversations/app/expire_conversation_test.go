package app

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

func TestExpireConversation_Execute(t *testing.T) {
	convID := uuid.NewV7()
	contactID := uuid.NewV7()
	now := time.Now()
	finishedAt := now.Add(-time.Hour)
	repoErr := errors.New("repository failure")

	tests := []struct {
		name       string
		setup      func(t *testing.T, repo *domain.MockConversationRepository) *domain.Conversation
		wantErr    error
		wantNoSave bool
		check      func(t *testing.T, conv *domain.Conversation)
	}{
		{
			name: "expires pending conversation",
			setup: func(t *testing.T, repo *domain.MockConversationRepository) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &contactID, nil)
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.ConversationStatusExpired, conv.Status())
				assert.NotNil(t, conv.FinishedAt())
			},
		},
		{
			name: "expires assigned conversation",
			setup: func(t *testing.T, repo *domain.MockConversationRepository) *domain.Conversation {
				t.Helper()
				agentID := uuid.NewV7()
				conv := buildConversation(t, convID, domain.ConversationStatusAssigned, &agentID, &contactID, nil)
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				repo.On("Save", mock.Anything, conv).Return(nil).Once()
				return conv
			},
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.ConversationStatusExpired, conv.Status())
			},
		},
		{
			name: "already finished is idempotent",
			setup: func(t *testing.T, repo *domain.MockConversationRepository) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusResolved, nil, &contactID, &finishedAt)
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(conv, nil).Once()
				return conv
			},
			wantNoSave: true,
			check: func(t *testing.T, conv *domain.Conversation) {
				t.Helper()
				assert.Equal(t, domain.ConversationStatusResolved, conv.Status())
			},
		},
		{
			name: "conversation not found",
			setup: func(t *testing.T, repo *domain.MockConversationRepository) *domain.Conversation {
				t.Helper()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(nil, nil).Once()
				return nil
			},
			wantErr: domain.ErrConversationNotFound,
		},
		{
			name: "find error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository) *domain.Conversation {
				t.Helper()
				repo.On("FindWithoutMessages", mock.Anything, convID).Return(nil, repoErr).Once()
				return nil
			},
			wantErr: repoErr,
		},
		{
			name: "save error",
			setup: func(t *testing.T, repo *domain.MockConversationRepository) *domain.Conversation {
				t.Helper()
				conv := buildConversation(t, convID, domain.ConversationStatusPending, nil, &contactID, nil)
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
			conv := tt.setup(t, repo)
			uc := NewExpireConversation(repo)

			// Act
			err := uc.Execute(context.Background(), ExpireConversationInput{ConversationID: convID, ExpiredAt: now})

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrExpiringConversation)
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
		})
	}
}
