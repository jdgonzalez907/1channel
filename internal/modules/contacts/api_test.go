package contacts

import (
	"context"
	"errors"
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/application"
	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewContactsAPI(t *testing.T) {
	// Arrange
	getOrCreateContact := &application.MockGetOrCreateContactByExternalID{}
	findContactByID := &application.MockFindContactByID{}

	// Act
	api := NewContactsAPI(getOrCreateContact, findContactByID)

	// Assert
	require.NotNil(t, api)
}

func TestContactsAPIGetOrCreateContactIDByExternalID(t *testing.T) {
	contactID := uuid.NewV7()
	externalContactID := "wa-contact-1"
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	existingContact, err := domain.NewContact(contactID, externalContactID, createdAt)
	require.NoError(t, err)

	tests := []struct {
		title             string
		setup             func(t *testing.T, m *application.MockGetOrCreateContactByExternalID)
		externalContactID string
		expected          uuid.UUID
		expectedError     string
	}{
		{
			title: "success - returns contact id when use case succeeds",
			setup: func(t *testing.T, m *application.MockGetOrCreateContactByExternalID) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(input application.GetOrCreateContactByExternalIDInput) bool {
					return input.ExternalContactID == externalContactID && input.ContactID != uuid.Nil() && !input.CreatedAt.IsZero()
				})).Return(existingContact, nil).Once()
			},
			externalContactID: externalContactID,
			expected:          contactID,
			expectedError:     "",
		},
		{
			title: "failure - returns nil id and use case error",
			setup: func(t *testing.T, m *application.MockGetOrCreateContactByExternalID) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("boom")).Once()
			},
			externalContactID: externalContactID,
			expected:          uuid.Nil(),
			expectedError:     "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			m := &application.MockGetOrCreateContactByExternalID{}
			f := &application.MockFindContactByID{}
			tt.setup(t, m)

			// Act
			got, err := NewContactsAPI(m, f).GetOrCreateContactIDByExternalID(context.Background(), tt.externalContactID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Equal(t, tt.expected, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
			m.AssertExpectations(t)
		})
	}
}
