package domain

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewPersonalInformation(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name     string
		document string
		optional func() (*string, *string, *string, *string, *string)
		wantErr  error
	}{
		{name: "document only", document: "12345678"},
		{name: "all fields", document: "ABC123", optional: func() (*string, *string, *string, *string, *string) {
			return strPtr("Juan"), strPtr("Perez"), strPtr("555"), strPtr("a@b.com"), strPtr("Calle 1")
		}},
		{name: "trims optional values", document: "12345678", optional: func() (*string, *string, *string, *string, *string) {
			return strPtr("  Juan  "), strPtr("  Perez  "), strPtr("  555  "), strPtr("  a@b.com  "), strPtr("  Calle 1  ")
		}},
		{name: "document at max length", document: strings.Repeat("9", 100)},
		{name: "empty document", document: "", wantErr: ErrIdentificationNumberEmpty},
		{name: "blank document", document: "   ", wantErr: ErrIdentificationNumberEmpty},
		{name: "document with dot", document: "12.345", wantErr: ErrIdentificationNumberInvalid},
		{name: "document with dash", document: "12-345", wantErr: ErrIdentificationNumberInvalid},
		{name: "document too long", document: strings.Repeat("9", 101), wantErr: ErrIdentificationNumberTooLong},
		{name: "first name too long", document: "123", optional: func() (*string, *string, *string, *string, *string) {
			return strPtr(strings.Repeat("a", 101)), nil, nil, nil, nil
		}, wantErr: ErrFirstNameTooLong},
		{name: "last name too long", document: "123", optional: func() (*string, *string, *string, *string, *string) {
			return nil, strPtr(strings.Repeat("a", 101)), nil, nil, nil
		}, wantErr: ErrLastNameTooLong},
		{name: "phone number too long", document: "123", optional: func() (*string, *string, *string, *string, *string) {
			return nil, nil, strPtr(strings.Repeat("1", 101)), nil, nil
		}, wantErr: ErrPhoneNumberTooLong},
		{name: "email too long", document: "123", optional: func() (*string, *string, *string, *string, *string) {
			return nil, nil, nil, strPtr(strings.Repeat("a", 255)), nil
		}, wantErr: ErrEmailTooLong},
		{name: "email at max length", document: "123", optional: func() (*string, *string, *string, *string, *string) {
			return nil, nil, nil, strPtr(strings.Repeat("a", 254)), nil
		}},
		{name: "address too long", document: "123", optional: func() (*string, *string, *string, *string, *string) {
			return nil, nil, nil, nil, strPtr(strings.Repeat("a", 255))
		}, wantErr: ErrAddressTooLong},
		{name: "optional length counted by graphemes", document: "123", optional: func() (*string, *string, *string, *string, *string) {
			return strPtr(strings.Repeat("e\u0301", 100)), nil, nil, nil, nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var firstName, lastName, phoneNumber, email, address *string
			if tt.optional != nil {
				firstName, lastName, phoneNumber, email, address = tt.optional()
			}

			// Act
			got, err := NewPersonalInformation(id, tt.document, firstName, lastName, phoneNumber, email, address, now, now)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)
		})
	}

	t.Run("nil identifier", func(t *testing.T) {
		// Act
		got, err := NewPersonalInformation(uuid.Nil(), "123", nil, nil, nil, nil, nil, now, now)

		// Assert
		assert.ErrorIs(t, err, ErrPersonalInformationInvalidID)
		assert.Nil(t, got)
	})

	t.Run("blank optional values become nil", func(t *testing.T) {
		// Act
		got, err := NewPersonalInformation(id, "12345678", strPtr("  "), nil, strPtr(""), nil, nil, now, now)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, got.FirstName())
		assert.Nil(t, got.PhoneNumber())
	})
}

func TestPersonalInformationUpdatePersonalData(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()
	later := now.Add(time.Hour)

	t.Run("replaces the data and moves updated_at", func(t *testing.T) {
		// Arrange
		personalInformation, err := NewPersonalInformation(id, "12345678", strPtr("Juan"), nil, nil, nil, nil, now, now)
		assert.NoError(t, err)

		// Act
		err = personalInformation.UpdatePersonalData(strPtr("Pedro"), strPtr("Gomez"), strPtr("555"), strPtr("p@b.com"), strPtr("Calle 2"), later)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, strPtr("Pedro"), personalInformation.FirstName())
		assert.Equal(t, strPtr("Gomez"), personalInformation.LastName())
		assert.Equal(t, later, personalInformation.UpdatedAt())
		assert.Equal(t, now, personalInformation.CreatedAt())
	})

	t.Run("empty values clear the data", func(t *testing.T) {
		// Arrange
		personalInformation, err := NewPersonalInformation(id, "12345678", strPtr("Juan"), strPtr("Perez"), nil, nil, nil, now, now)
		assert.NoError(t, err)

		// Act
		err = personalInformation.UpdatePersonalData(nil, nil, nil, nil, nil, later)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, personalInformation.FirstName())
		assert.Nil(t, personalInformation.LastName())
	})

	t.Run("too long optional value is rejected", func(t *testing.T) {
		// Arrange
		personalInformation, err := NewPersonalInformation(id, "12345678", nil, nil, nil, nil, nil, now, now)
		assert.NoError(t, err)

		// Act
		err = personalInformation.UpdatePersonalData(strPtr(strings.Repeat("a", 101)), nil, nil, nil, nil, later)

		// Assert
		assert.ErrorIs(t, err, ErrFirstNameTooLong)
	})
}

func TestRehydratePersonalInformation(t *testing.T) {
	id := uuid.NewV7()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	// Act
	got := RehydratePersonalInformation(id, "", nil, nil, nil, nil, nil, createdAt, updatedAt)

	// Assert
	assert.Equal(t, id, got.ID())
	assert.Equal(t, "", got.IdentificationNumber())
	assert.Nil(t, got.FirstName())
	assert.Equal(t, createdAt, got.CreatedAt())
	assert.Equal(t, updatedAt, got.UpdatedAt())
}
