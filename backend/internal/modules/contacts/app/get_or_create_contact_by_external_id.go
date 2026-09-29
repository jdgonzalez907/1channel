package app

import (
	"context"
	"errors"
	"strings"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

var ErrGettingOrCreatingContactByExternalID = errors.New("getting or creating contact by external id")

type GetOrCreateContactByExternalIDInput struct {
	ContactID         uuid.UUID
	ExternalContactID string
	DisplayName       *string
	CreatedAt         time.Time
}

type GetOrCreateContactByExternalID interface {
	Execute(ctx context.Context, input GetOrCreateContactByExternalIDInput) (*domain.Contact, error)
}

type getOrCreateContactByExternalID struct {
	contactRepository domain.ContactRepository
}

func NewGetOrCreateContactByExternalID(contactRepository domain.ContactRepository) GetOrCreateContactByExternalID {
	return &getOrCreateContactByExternalID{contactRepository: contactRepository}
}

func (uc *getOrCreateContactByExternalID) Execute(ctx context.Context, input GetOrCreateContactByExternalIDInput) (*domain.Contact, error) {
	contact, err := uc.contactRepository.FindByExternalContactID(ctx, input.ExternalContactID)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	if contact != nil {
		return uc.updateDisplayName(ctx, contact, input.DisplayName)
	}

	contact, err = domain.NewContact(input.ContactID, input.ExternalContactID, input.CreatedAt)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	contact.AssignDisplayName(input.DisplayName)

	if err := uc.contactRepository.Save(ctx, contact); err != nil {
		return nil, uc.joinErr(err)
	}

	return contact, nil
}

func (uc *getOrCreateContactByExternalID) updateDisplayName(ctx context.Context, contact *domain.Contact, displayName *string) (*domain.Contact, error) {
	if displayName == nil || strings.TrimSpace(*displayName) == "" {
		return contact, nil
	}

	contact.AssignDisplayName(displayName)

	if err := uc.contactRepository.Save(ctx, contact); err != nil {
		return nil, uc.joinErr(err)
	}

	return contact, nil
}

func (uc *getOrCreateContactByExternalID) joinErr(errs ...error) error {
	return errors.Join(ErrGettingOrCreatingContactByExternalID, errors.Join(errs...))
}
