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

func TestRehydrateContact(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name              string
		externalContactID string
	}{
		{
			name:              "with external contact id",
			externalContactID: "5491112345678",
		},
		{
			name:              "empty external contact id",
			externalContactID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			contact := RehydrateContact(id, tt.externalContactID, nil, nil, now)

			// Assert
			assert.Equal(t, id, contact.ID())
			assert.Equal(t, tt.externalContactID, contact.ExternalContactID())
			assert.Equal(t, now, contact.CreatedAt())
			assert.Nil(t, contact.DisplayName())
			assert.Nil(t, contact.PersonalInformationID())
		})
	}

	t.Run("with display name and personal information", func(t *testing.T) {
		// Arrange
		displayName := "Juan Perez"
		personalInformationID := uuid.NewV7()

		// Act
		contact := RehydrateContact(id, "54911", &displayName, &personalInformationID, now)

		// Assert
		assert.Equal(t, &displayName, contact.DisplayName())
		assert.Equal(t, &personalInformationID, contact.PersonalInformationID())
	})
}

func TestContactAssignDisplayName(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name        string
		displayName *string
		want        *string
	}{
		{name: "sets trimmed value", displayName: strPtr("  Juan  "), want: strPtr("Juan")},
		{name: "ignores nil", displayName: nil, want: nil},
		{name: "ignores empty", displayName: strPtr("   "), want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			contact, err := NewContact(id, "54911", now)
			assert.NoError(t, err)

			// Act
			contact.AssignDisplayName(tt.displayName)

			// Assert
			assert.Equal(t, tt.want, contact.DisplayName())
		})
	}
}

func TestContactAssociatePersonalInformation(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()
	personalInformationID := uuid.NewV7()

	// Arrange
	contact, err := NewContact(id, "54911", now)
	assert.NoError(t, err)

	// Act
	contact.AssociatePersonalInformation(personalInformationID)

	// Assert
	assert.Equal(t, &personalInformationID, contact.PersonalInformationID())
}

func strPtr(value string) *string {
	return &value
}
