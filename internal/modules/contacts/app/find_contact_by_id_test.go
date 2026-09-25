package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindContactByID_Execute(t *testing.T) {
	contactID := uuid.NewV7()
	now := time.Now()
	contact, err := domain.NewContact(contactID, "5491112345678", now)
	assert.NoError(t, err)
	repoErr := errors.New("repository failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *domain.MockContactRepository)
		input   FindContactByIDInput
		want    *domain.Contact
		wantErr error
	}{
		{
			name: "contact found",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, contactID).Return(contact, nil).Once()
			},
			input: FindContactByIDInput{ContactID: contactID},
			want:  contact,
		},
		{
			name: "contact not found",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, contactID).Return(nil, nil).Once()
			},
			input:   FindContactByIDInput{ContactID: contactID},
			wantErr: domain.ErrContactNotFound,
		},
		{
			name: "repository error",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByID", mock.Anything, contactID).Return(nil, repoErr).Once()
			},
			input:   FindContactByIDInput{ContactID: contactID},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &domain.MockContactRepository{}
			tt.setup(t, repo)
			uc := NewFindContactByID(repo)

			// Act
			got, err := uc.Execute(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrFindingContactByID)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			repo.AssertExpectations(t)
		})
	}
}
