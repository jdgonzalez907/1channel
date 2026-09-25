package domain

import (
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewContact(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name              string
		id                uuid.UUID
		externalContactID string
		createdAt         time.Time
		wantErr           error
	}{
		{
			name:              "creates contact with assigned fields",
			id:                id,
			externalContactID: "5491112345678",
			createdAt:         now,
		},
		{
			name:              "empty external contact id",
			id:                id,
			externalContactID: "",
			createdAt:         now,
			wantErr:           ErrExternalContactIDEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			contact, err := NewContact(tt.id, tt.externalContactID, tt.createdAt)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, contact)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.id, contact.ID())
			assert.Equal(t, tt.externalContactID, contact.ExternalContactID())
			assert.Equal(t, tt.createdAt, contact.CreatedAt())
		})
	}
}
