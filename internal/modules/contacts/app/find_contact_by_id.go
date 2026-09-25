package app

import (
	"context"
	"errors"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

var ErrFindingContactByID = errors.New("finding contact by id")

type FindContactByIDInput struct {
	ContactID uuid.UUID
}

type FindContactByID interface {
	Execute(ctx context.Context, input FindContactByIDInput) (*domain.Contact, error)
}

type findContactByID struct {
	contactRepository domain.ContactRepository
}

func NewFindContactByID(contactRepository domain.ContactRepository) FindContactByID {
	return &findContactByID{contactRepository: contactRepository}
}

func (uc *findContactByID) Execute(ctx context.Context, input FindContactByIDInput) (*domain.Contact, error) {
	contact, err := uc.contactRepository.FindByID(ctx, input.ContactID)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	if contact == nil {
		return nil, uc.joinErr(domain.ErrContactNotFound)
	}

	return contact, nil
}

func (uc *findContactByID) joinErr(errs ...error) error {
	return errors.Join(ErrFindingContactByID, errors.Join(errs...))
}
