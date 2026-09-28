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

func TestFindUserByID_Execute(t *testing.T) {
	userID := uuid.NewV7()
	now := time.Now()
	user, err := domain.NewUser(userID, now)
	assert.NoError(t, err)
	repoErr := errors.New("repository failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *domain.MockUserRepository)
		input   FindUserByIDInput
		want    *domain.User
		wantErr error
	}{
		{
			name: "user found",
			setup: func(t *testing.T, repo *domain.MockUserRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, userID).Return(user, nil).Once()
			},
			input: FindUserByIDInput{ID: userID},
			want:  user,
		},
		{
			name: "user not found",
			setup: func(t *testing.T, repo *domain.MockUserRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, userID).Return(nil, nil).Once()
			},
			input:   FindUserByIDInput{ID: userID},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name: "repository error",
			setup: func(t *testing.T, repo *domain.MockUserRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, userID).Return(nil, repoErr).Once()
			},
			input:   FindUserByIDInput{ID: userID},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &domain.MockUserRepository{}
			tt.setup(t, repo)
			uc := NewFindUserByID(repo)

			// Act
			got, err := uc.Execute(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrFindingUserByID)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			repo.AssertExpectations(t)
		})
	}
}
