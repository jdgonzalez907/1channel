package application

import (
	"context"
	"errors"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

type (
	GetOrCreateContactByExternalIDInput struct {
		ContactID         uuid.UUID
		ExternalContactID string
	}

	GetOrCreateContactByExternalID interface {
		Execute(ctx context.Context, input GetOrCreateContactByExternalIDInput) (*domain.Contact, error)
	}

	getOrCreateContactByExternalID struct {
		contactRepository domain.ContactRepository
	}
)

func NewGetOrCreateContactByExternalID(contactRepository domain.ContactRepository) GetOrCreateContactByExternalID {
	return &getOrCreateContactByExternalID{contactRepository}
}

func (uc *getOrCreateContactByExternalID) Execute(ctx context.Context, input GetOrCreateContactByExternalIDInput) (*domain.Contact, error) {
	contact, err := uc.contactRepository.FindByExternalContactID(ctx, input.ExternalContactID)
	if err != nil {
		return nil, uc.wrapError(err)
	}

	if contact != nil {
		return contact, nil
	}

	contact, err = domain.NewContact(input.ContactID, input.ExternalContactID)
	if err != nil {
		return nil, uc.wrapError(err)
	}

	err = uc.contactRepository.Save(ctx, contact)
	if err != nil {
		return nil, uc.wrapError(err)
	}

	return contact, nil
}

func (uc *getOrCreateContactByExternalID) wrapError(err error) error {
	return errors.Join(errors.New("error getting or creating contact by external id"), err)
}
