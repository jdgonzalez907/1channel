package app

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

func TestCreateUser_Execute(t *testing.T) {
	userID := uuid.NewV7()
	now := time.Now()
	repoErr := errors.New("repository failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *domain.MockUserRepository)
		input   CreateUserInput
		want    uuid.UUID
		wantErr error
	}{
		{
			name: "creates user",
			setup: func(t *testing.T, repo *domain.MockUserRepository) {
				t.Helper()
				repo.On("Save", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.ID() == userID
				})).Return(nil).Once()
			},
			input: CreateUserInput{ID: userID, CreatedAt: now},
			want:  userID,
		},
		{
			name:    "fails with nil id",
			setup:   func(t *testing.T, repo *domain.MockUserRepository) {},
			input:   CreateUserInput{ID: uuid.Nil(), CreatedAt: now},
			wantErr: domain.ErrUserInvalidID,
		},
		{
			name: "repository error",
			setup: func(t *testing.T, repo *domain.MockUserRepository) {
				t.Helper()
				repo.On("Save", mock.Anything, mock.Anything).Return(repoErr).Once()
			},
			input:   CreateUserInput{ID: userID, CreatedAt: now},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &domain.MockUserRepository{}
			tt.setup(t, repo)
			uc := NewCreateUser(repo)

			// Act
			got, err := uc.Execute(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrCreatingUser)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.ID())
			}
			repo.AssertExpectations(t)
		})
	}
}
