package app

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

func strPtr(value string) *string {
	return &value
}

func TestRegisterContactPersonalInformation_Execute(t *testing.T) {
	contactID := uuid.NewV7()
	personalInformationID := uuid.NewV7()
	now := time.Now()
	repoErr := errors.New("repository failure")

	t.Run("registers a new person and associates the contact", func(t *testing.T) {
		// Arrange
		personalInformationRepo := &domain.MockPersonalInformationRepository{}
		contactRepo := &domain.MockContactRepository{}
		contact, err := domain.NewContact(contactID, "54911", now)
		assert.NoError(t, err)

		find := personalInformationRepo.On("FindByIdentificationNumber", mock.Anything, "12345678").Return(nil, nil).Once()
		savePI := personalInformationRepo.On("Save", mock.Anything, mock.MatchedBy(func(pi *domain.PersonalInformation) bool {
			return pi.IdentificationNumber() == "12345678" && pi.FirstName() != nil && *pi.FirstName() == "Juan"
		})).Return(nil).Once()
		findContact := contactRepo.On("FindByID", mock.Anything, contactID).Return(contact, nil).Once()
		saveContact := contactRepo.On("Save", mock.Anything, mock.MatchedBy(func(c *domain.Contact) bool {
			return c.PersonalInformationID() != nil && *c.PersonalInformationID() == personalInformationID
		})).Return(nil).Once()
		mock.InOrder(find, savePI, findContact, saveContact)

		uc := NewRegisterContactPersonalInformation(personalInformationRepo, contactRepo)

		// Act
		got, err := uc.Execute(context.Background(), RegisterContactPersonalInformationInput{
			ContactID:             contactID,
			PersonalInformationID: personalInformationID,
			IdentificationNumber:  "12345678",
			FirstName:             strPtr("Juan"),
			At:                    now,
		})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, personalInformationID, got.ID())
		personalInformationRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})

	t.Run("updates the existing person and keeps its identifier", func(t *testing.T) {
		// Arrange
		personalInformationRepo := &domain.MockPersonalInformationRepository{}
		contactRepo := &domain.MockContactRepository{}
		existing := domain.RehydratePersonalInformation(personalInformationID, "12345678", strPtr("Juan"), strPtr("Perez"), nil, nil, nil, now.Add(-time.Hour), now.Add(-time.Hour))
		contact, err := domain.NewContact(contactID, "54911", now)
		assert.NoError(t, err)

		personalInformationRepo.On("FindByIdentificationNumber", mock.Anything, "12345678").Return(existing, nil).Once()
		personalInformationRepo.On("Save", mock.Anything, mock.MatchedBy(func(pi *domain.PersonalInformation) bool {
			return pi.FirstName() != nil && *pi.FirstName() == "Pedro"
		})).Return(nil).Once()
		contactRepo.On("FindByID", mock.Anything, contactID).Return(contact, nil).Once()
		contactRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

		uc := NewRegisterContactPersonalInformation(personalInformationRepo, contactRepo)

		// Act
		got, err := uc.Execute(context.Background(), RegisterContactPersonalInformationInput{
			ContactID:             contactID,
			PersonalInformationID: uuid.NewV7(),
			IdentificationNumber:  "12345678",
			FirstName:             strPtr("Pedro"),
			At:                    now,
		})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, personalInformationID, got.ID())
		assert.Equal(t, "Pedro", *got.FirstName())
		personalInformationRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})

	t.Run("empty values become nil", func(t *testing.T) {
		// Arrange
		personalInformationRepo := &domain.MockPersonalInformationRepository{}
		contactRepo := &domain.MockContactRepository{}
		contact, err := domain.NewContact(contactID, "54911", now)
		assert.NoError(t, err)

		personalInformationRepo.On("FindByIdentificationNumber", mock.Anything, "12345678").Return(nil, nil).Once()
		personalInformationRepo.On("Save", mock.Anything, mock.MatchedBy(func(pi *domain.PersonalInformation) bool {
			return pi.FirstName() == nil && pi.LastName() == nil
		})).Return(nil).Once()
		contactRepo.On("FindByID", mock.Anything, contactID).Return(contact, nil).Once()
		contactRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()

		uc := NewRegisterContactPersonalInformation(personalInformationRepo, contactRepo)

		// Act
		_, err = uc.Execute(context.Background(), RegisterContactPersonalInformationInput{
			ContactID:             contactID,
			PersonalInformationID: uuid.NewV7(),
			IdentificationNumber:  "12345678",
			FirstName:             strPtr("   "),
			At:                    now,
		})

		// Assert
		assert.NoError(t, err)
		personalInformationRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})

	t.Run("invalid document is rejected before saving", func(t *testing.T) {
		// Arrange
		personalInformationRepo := &domain.MockPersonalInformationRepository{}
		contactRepo := &domain.MockContactRepository{}
		personalInformationRepo.On("FindByIdentificationNumber", mock.Anything, "12.345").Return(nil, nil).Once()

		uc := NewRegisterContactPersonalInformation(personalInformationRepo, contactRepo)

		// Act
		got, err := uc.Execute(context.Background(), RegisterContactPersonalInformationInput{
			ContactID:             contactID,
			PersonalInformationID: uuid.NewV7(),
			IdentificationNumber:  "12.345",
			At:                    now,
		})

		// Assert
		assert.ErrorIs(t, err, domain.ErrIdentificationNumberInvalid)
		assert.ErrorIs(t, err, ErrRegisteringContactPersonalInformation)
		assert.Nil(t, got)
		personalInformationRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})

	t.Run("contact not found", func(t *testing.T) {
		// Arrange
		personalInformationRepo := &domain.MockPersonalInformationRepository{}
		contactRepo := &domain.MockContactRepository{}
		personalInformationRepo.On("FindByIdentificationNumber", mock.Anything, "12345678").Return(nil, nil).Once()
		personalInformationRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
		contactRepo.On("FindByID", mock.Anything, contactID).Return(nil, nil).Once()

		uc := NewRegisterContactPersonalInformation(personalInformationRepo, contactRepo)

		// Act
		got, err := uc.Execute(context.Background(), RegisterContactPersonalInformationInput{
			ContactID:             contactID,
			PersonalInformationID: uuid.NewV7(),
			IdentificationNumber:  "12345678",
			At:                    now,
		})

		// Assert
		assert.ErrorIs(t, err, domain.ErrContactNotFound)
		assert.Nil(t, got)
		personalInformationRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})

	t.Run("personal information save error is propagated", func(t *testing.T) {
		// Arrange
		personalInformationRepo := &domain.MockPersonalInformationRepository{}
		contactRepo := &domain.MockContactRepository{}
		personalInformationRepo.On("FindByIdentificationNumber", mock.Anything, "12345678").Return(nil, nil).Once()
		personalInformationRepo.On("Save", mock.Anything, mock.Anything).Return(repoErr).Once()

		uc := NewRegisterContactPersonalInformation(personalInformationRepo, contactRepo)

		// Act
		got, err := uc.Execute(context.Background(), RegisterContactPersonalInformationInput{
			ContactID:             contactID,
			PersonalInformationID: uuid.NewV7(),
			IdentificationNumber:  "12345678",
			At:                    now,
		})

		// Assert
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
		personalInformationRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})

	t.Run("contact save error is propagated", func(t *testing.T) {
		// Arrange
		personalInformationRepo := &domain.MockPersonalInformationRepository{}
		contactRepo := &domain.MockContactRepository{}
		contact, err := domain.NewContact(contactID, "54911", now)
		assert.NoError(t, err)

		personalInformationRepo.On("FindByIdentificationNumber", mock.Anything, "12345678").Return(nil, nil).Once()
		personalInformationRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
		contactRepo.On("FindByID", mock.Anything, contactID).Return(contact, nil).Once()
		contactRepo.On("Save", mock.Anything, mock.Anything).Return(repoErr).Once()

		uc := NewRegisterContactPersonalInformation(personalInformationRepo, contactRepo)

		// Act
		got, err := uc.Execute(context.Background(), RegisterContactPersonalInformationInput{
			ContactID:             contactID,
			PersonalInformationID: uuid.NewV7(),
			IdentificationNumber:  "12345678",
			At:                    now,
		})

		// Assert
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
		personalInformationRepo.AssertExpectations(t)
		contactRepo.AssertExpectations(t)
	})
}
