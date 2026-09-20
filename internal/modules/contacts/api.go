package contacts

import (
	"context"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/application"
)

type ContactsAPI interface {
	GetOrCreateContactByExternalID(ctx context.Context, externalContactID string) (uuid.UUID, error)
}

type contactsAPI struct {
	getOrCreateContactByExternalID application.GetOrCreateContactByExternalID
}

func NewContactsAPI(getOrCreateContactByExternalID application.GetOrCreateContactByExternalID) ContactsAPI {
	return &contactsAPI{getOrCreateContactByExternalID}
}

func (a *contactsAPI) GetOrCreateContactByExternalID(ctx context.Context, externalContactID string) (uuid.UUID, error) {
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
