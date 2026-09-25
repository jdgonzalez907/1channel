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

func TestGetOrCreateContactByExternalID_Execute(t *testing.T) {
	contactID := uuid.NewV7()
	externalID := "5491112345678"
	now := time.Now()
	existing, err := domain.NewContact(contactID, externalID, now)
	assert.NoError(t, err)
	repoErr := errors.New("repository failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *domain.MockContactRepository)
		input   GetOrCreateContactByExternalIDInput
		want    *domain.Contact
		wantErr error
	}{
		{
			name: "existing contact",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByExternalContactID", mock.Anything, externalID).Return(existing, nil).Once()
			},
			input: GetOrCreateContactByExternalIDInput{ContactID: contactID, ExternalContactID: externalID, CreatedAt: now},
			want:  existing,
		},
		{
			name: "new contact",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByExternalContactID", mock.Anything, externalID).Return(nil, nil).Once()
				repo.On("Save", mock.Anything, mock.MatchedBy(func(c *domain.Contact) bool {
					return c.ExternalContactID() == externalID
				})).Return(nil).Once()
			},
			input: GetOrCreateContactByExternalIDInput{ContactID: contactID, ExternalContactID: externalID, CreatedAt: now},
		},
		{
			name: "empty external id",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByExternalContactID", mock.Anything, "").Return(nil, nil).Once()
			},
			input:   GetOrCreateContactByExternalIDInput{ContactID: contactID, ExternalContactID: "", CreatedAt: now},
			wantErr: domain.ErrExternalContactIDEmpty,
		},
		{
			name: "find error",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByExternalContactID", mock.Anything, externalID).Return(nil, repoErr).Once()
			},
			input:   GetOrCreateContactByExternalIDInput{ContactID: contactID, ExternalContactID: externalID, CreatedAt: now},
			wantErr: repoErr,
		},
		{
			name: "save error",
			setup: func(t *testing.T, repo *domain.MockContactRepository) {
				t.Helper()
				repo.On("FindByExternalContactID", mock.Anything, externalID).Return(nil, nil).Once()
				repo.On("Save", mock.Anything, mock.Anything).Return(repoErr).Once()
			},
			input:   GetOrCreateContactByExternalIDInput{ContactID: contactID, ExternalContactID: externalID, CreatedAt: now},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &domain.MockContactRepository{}
			tt.setup(t, repo)
			uc := NewGetOrCreateContactByExternalID(repo)

			// Act
			got, err := uc.Execute(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.ErrorIs(t, err, ErrGettingOrCreatingContactByExternalID)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.want != nil {
					assert.Equal(t, tt.want, got)
				}
			}
			repo.AssertExpectations(t)
		})
	}
}
