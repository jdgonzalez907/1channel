package domain

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContact(t *testing.T) {
	id := uuid.NewV7()

	tests := []struct {
		title             string
		id                uuid.UUID
		externalContactID string
		expectedError     string
	}{
		{
			title:             "success - builds contact with external id",
			id:                id,
			externalContactID: "wa-contact-1",
			expectedError:     "",
		},
		{
			title:             "failure - rejects empty external id boundary",
			id:                id,
			externalContactID: "",
			expectedError:     ErrExternalContactIDEmpty.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			contact, err := NewContact(tt.id, tt.externalContactID)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				require.Nil(t, contact)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, contact)
			assert.Equal(t, tt.id, contact.ID())
			assert.Equal(t, tt.externalContactID, contact.ExternalContactID())
		})
	}
}
