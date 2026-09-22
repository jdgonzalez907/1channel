package contacts

import (
	"context"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/application"
)

type ContactsAPI interface {
	GetOrCreateContactIDByExternalID(ctx context.Context, externalContactID string) (uuid.UUID, error)
	FindExternalContactIDByContactID(ctx context.Context, contactID uuid.UUID) (string, error)
}

type contactsAPI struct {
	getOrCreateContactByExternalID application.GetOrCreateContactByExternalID
	findContactByID                application.FindContactByID
}

func NewContactsAPI(getOrCreateContactByExternalID application.GetOrCreateContactByExternalID, findContactByID application.FindContactByID) ContactsAPI {
	return &contactsAPI{getOrCreateContactByExternalID, findContactByID}
}

func (a *contactsAPI) GetOrCreateContactIDByExternalID(ctx context.Context, externalContactID string) (uuid.UUID, error) {
	contact, err := a.getOrCreateContactByExternalID.Execute(ctx, application.GetOrCreateContactByExternalIDInput{
		ContactID:         uuid.NewV7(),
		ExternalContactID: externalContactID,
		CreatedAt:         time.Now().UTC(),
	})
	if err != nil {
		return uuid.Nil(), err
	}

	return contact.ID(), nil
}

func (a *contactsAPI) FindExternalContactIDByContactID(ctx context.Context, contactID uuid.UUID) (string, error) {
	contact, err := a.findContactByID.Execute(ctx, application.FindContactByIDInput{
		ContactID: contactID,
	})
	if err != nil {
		return "", err
	}

	return contact.ExternalContactID(), nil
}
