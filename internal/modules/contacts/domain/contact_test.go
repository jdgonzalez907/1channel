package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContact(t *testing.T) {
	id := uuid.NewV7()
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		title             string
		id                uuid.UUID
		externalContactID string
		createdAt         time.Time
		expectedError     string
	}{
		{
			title:             "success - builds contact with external id",
			id:                id,
			externalContactID: "wa-contact-1",
			createdAt:         createdAt,
			expectedError:     "",
		},
		{
			title:             "failure - rejects empty external id boundary",
			id:                id,
			externalContactID: "",
			createdAt:         createdAt,
			expectedError:     ErrExternalContactIDEmpty.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			// (inputs are in the table)

			// Act
			contact, err := NewContact(tt.id, tt.externalContactID, tt.createdAt)

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
			assert.Equal(t, tt.createdAt, contact.CreatedAt())
		})
	}
}
