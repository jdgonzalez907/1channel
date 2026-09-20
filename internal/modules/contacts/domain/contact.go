package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrExternalContactIDEmpty = errors.New("external contact id is empty")
)

type Contact struct {
	id                uuid.UUID
	externalContactID string
	createdAt         time.Time
}

func NewContact(id uuid.UUID, externalContactID string, createdAt time.Time) (*Contact, error) {
	if externalContactID == "" {
		return nil, ErrExternalContactIDEmpty
	}

	return &Contact{id, externalContactID, createdAt}, nil
}

func (c *Contact) ID() uuid.UUID             { return c.id }
func (c *Contact) ExternalContactID() string { return c.externalContactID }
func (c *Contact) CreatedAt() time.Time      { return c.createdAt }
