package app

import (
	"context"
	"errors"
	"strings"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

var ErrRegisteringContactPersonalInformation = errors.New("registering contact personal information")

type RegisterContactPersonalInformationInput struct {
	ContactID             uuid.UUID
	PersonalInformationID uuid.UUID
	IdentificationNumber  string
	FirstName             *string
	LastName              *string
	PhoneNumber           *string
	Email                 *string
	Address               *string
	At                    time.Time
}

type RegisterContactPersonalInformation interface {
	Execute(ctx context.Context, input RegisterContactPersonalInformationInput) (*domain.PersonalInformation, error)
}

type registerContactPersonalInformation struct {
	personalInformationRepository domain.PersonalInformationRepository
	contactRepository             domain.ContactRepository
}

func NewRegisterContactPersonalInformation(
	personalInformationRepository domain.PersonalInformationRepository,
	contactRepository domain.ContactRepository,
) RegisterContactPersonalInformation {
	return &registerContactPersonalInformation{
		personalInformationRepository: personalInformationRepository,
		contactRepository:             contactRepository,
	}
}

func (uc *registerContactPersonalInformation) Execute(ctx context.Context, input RegisterContactPersonalInformationInput) (*domain.PersonalInformation, error) {
	identificationNumber := strings.TrimSpace(input.IdentificationNumber)

	personalInformation, err := uc.personalInformationRepository.FindByIdentificationNumber(ctx, identificationNumber)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	if personalInformation != nil {
		if err := personalInformation.UpdatePersonalData(input.FirstName, input.LastName, input.PhoneNumber, input.Email, input.Address, input.At); err != nil {
			return nil, uc.joinErr(err)
		}

		if err := uc.personalInformationRepository.Save(ctx, personalInformation); err != nil {
			return nil, uc.joinErr(err)
		}
	} else {
		personalInformation, err = domain.NewPersonalInformation(
			input.PersonalInformationID,
			identificationNumber,
			input.FirstName,
			input.LastName,
			input.PhoneNumber,
			input.Email,
			input.Address,
			input.At,
			input.At,
		)
		if err != nil {
			return nil, uc.joinErr(err)
		}

		if err := uc.personalInformationRepository.Save(ctx, personalInformation); err != nil {
			return nil, uc.joinErr(err)
		}
	}

	contact, err := uc.contactRepository.FindByID(ctx, input.ContactID)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	if contact == nil {
		return nil, uc.joinErr(domain.ErrContactNotFound)
	}

	contact.AssociatePersonalInformation(personalInformation.ID())

	if err := uc.contactRepository.Save(ctx, contact); err != nil {
		return nil, uc.joinErr(err)
	}

	return personalInformation, nil
}

func (uc *registerContactPersonalInformation) joinErr(errs ...error) error {
	return errors.Join(ErrRegisteringContactPersonalInformation, errors.Join(errs...))
}
