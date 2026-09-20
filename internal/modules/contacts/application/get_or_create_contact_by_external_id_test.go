package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"uuid"
)

func TestNewGetOrCreateContactByExternalID(t *testing.T) {
	// Arrange
	repository := &domain.MockContactRepository{}

	// Act
	useCase := NewGetOrCreateContactByExternalID(repository)

	// Assert
	require.NotNil(t, useCase)
}

func TestGetOrCreateContactByExternalIDExecute(t *testing.T) {
	contactID := uuid.NewV7()
	externalContactID := "wa-contact-1"
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	existingContact, err := domain.NewContact(contactID, externalContactID, createdAt)
	require.NoError(t, err)

	input := GetOrCreateContactByExternalIDInput{
		ContactID:         contactID,
		ExternalContactID: externalContactID,
		CreatedAt:         createdAt,
	}

	tests := []struct {
		title         string
		setup         func(t *testing.T, m *domain.MockContactRepository)
		input         GetOrCreateContactByExternalIDInput
		expectedID    uuid.UUID
		expectedError string
	}{
		{
			title: "success - creates and saves contact when none exists",
			setup: func(t *testing.T, m *domain.MockContactRepository) {
				t.Helper()
				m.On("FindByExternalContactID", mock.Anything, externalContactID).Return(nil, nil).Once()
				m.On("Save", mock.Anything, mock.MatchedBy(func(saved *domain.Contact) bool {
					return saved != nil && saved.ID() == contactID && saved.ExternalContactID() == externalContactID && saved.CreatedAt().Equal(createdAt)
				})).Return(nil).Once()
			},
			input:         input,
			expectedID:    contactID,
			expectedError: "",
		},
		{
			title: "success - returns existing contact when already created",
			setup: func(t *testing.T, m *domain.MockContactRepository) {
				t.Helper()
				m.On("FindByExternalContactID", mock.Anything, externalContactID).Return(existingContact, nil).Once()
			},
			input:         input,
			expectedID:    contactID,
			expectedError: "",
		},
		{
			title: "failure - wraps repository find error",
			setup: func(t *testing.T, m *domain.MockContactRepository) {
				t.Helper()
				m.On("FindByExternalContactID", mock.Anything, externalContactID).Return(nil, errors.New("db connection lost")).Once()
			},
			input:         input,
			expectedID:    uuid.Nil(),
			expectedError: "error getting or creating contact by external id\ndb connection lost",
		},
		{
			title: "failure - wraps empty external id validation error",
			setup: func(t *testing.T, m *domain.MockContactRepository) {
				t.Helper()
				m.On("FindByExternalContactID", mock.Anything, "").Return(nil, nil).Once()
			},
			input: GetOrCreateContactByExternalIDInput{
				ContactID:         contactID,
				ExternalContactID: "",
				CreatedAt:         createdAt,
			},
			expectedID:    uuid.Nil(),
			expectedError: "error getting or creating contact by external id\n" + domain.ErrExternalContactIDEmpty.Error(),
		},
		{
			title: "failure - wraps repository save error",
			setup: func(t *testing.T, m *domain.MockContactRepository) {
				t.Helper()
				m.On("FindByExternalContactID", mock.Anything, externalContactID).Return(nil, nil).Once()
				m.On("Save", mock.Anything, mock.Anything).Return(errors.New("save failed")).Once()
			},
			input:         input,
			expectedID:    uuid.Nil(),
			expectedError: "error getting or creating contact by external id\nsave failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &domain.MockContactRepository{}
			tt.setup(t, m)

			// Act
			contact, err := NewGetOrCreateContactByExternalID(m).Execute(context.Background(), tt.input)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, contact)
			} else {
				require.NoError(t, err)
				require.NotNil(t, contact)
				assert.Equal(t, tt.expectedID, contact.ID())
			}
			m.AssertExpectations(t)
		})
	}
}
