package domain

import (
	"errors"
	"uuid"
)

var (
	ErrExternalContactIDEmpty = errors.New("external contact id is empty")
)

type Contact struct {
	id                uuid.UUID
	externalContactID string
}

func NewContact(id uuid.UUID, externalContactID string) (*Contact, error) {
	if externalContactID == "" {
		return nil, ErrExternalContactIDEmpty
	}

	return &Contact{id, externalContactID}, nil
}

func (c *Contact) ID() uuid.UUID             { return c.id }
func (c *Contact) ExternalContactID() string { return c.externalContactID }
