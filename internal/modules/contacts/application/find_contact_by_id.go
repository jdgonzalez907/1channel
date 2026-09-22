package application

import (
	"context"
	"errors"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

type (
	FindContactByIDInput struct {
		ContactID uuid.UUID
	}

	FindContactByID interface {
		Execute(ctx context.Context, input FindContactByIDInput) (*domain.Contact, error)
	}

	findContactByID struct {
		contactRepository domain.ContactRepository
	}
)

func NewFindContactByID(contactRepository domain.ContactRepository) FindContactByID {
	return &findContactByID{contactRepository}
}

func (uc *findContactByID) Execute(ctx context.Context, input FindContactByIDInput) (*domain.Contact, error) {
	contact, err := uc.contactRepository.FindByID(ctx, input.ContactID)
	if err != nil {
		return nil, uc.wrapError(err)
	}

	if contact == nil {
		return nil, uc.wrapError(domain.ErrContactNotFound)
	}

	return contact, nil
}

func (uc *findContactByID) wrapError(err error) error {
	return errors.Join(errors.New("error finding contact by id"), err)
}
