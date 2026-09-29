package domain

import (
	"errors"
	"strings"
	"time"
	"uuid"
)

var (
	ErrExternalContactIDEmpty = errors.New("external contact id is empty")
	ErrContactNotFound        = errors.New("contact not found")
)

type Contact struct {
	id                    uuid.UUID
	externalContactID     string
	displayName           *string
	personalInformationID *uuid.UUID
	createdAt             time.Time
}

func NewContact(id uuid.UUID, externalContactID string, createdAt time.Time) (*Contact, error) {
	if externalContactID == "" {
		return nil, ErrExternalContactIDEmpty
	}

	return &Contact{id: id, externalContactID: externalContactID, createdAt: createdAt}, nil
}

func RehydrateContact(
	id uuid.UUID,
	externalContactID string,
	displayName *string,
	personalInformationID *uuid.UUID,
	createdAt time.Time,
) *Contact {
	return &Contact{
		id:                    id,
		externalContactID:     externalContactID,
		displayName:           displayName,
		personalInformationID: personalInformationID,
		createdAt:             createdAt,
	}
}

func (c *Contact) ID() uuid.UUID                     { return c.id }
func (c *Contact) ExternalContactID() string         { return c.externalContactID }
func (c *Contact) DisplayName() *string              { return c.displayName }
func (c *Contact) PersonalInformationID() *uuid.UUID { return c.personalInformationID }
func (c *Contact) CreatedAt() time.Time              { return c.createdAt }

func (c *Contact) AssignDisplayName(displayName *string) {
	if displayName == nil {
		return
	}

	value := strings.TrimSpace(*displayName)
	if value == "" {
		return
	}

	c.displayName = &value
}

func (c *Contact) AssociatePersonalInformation(personalInformationID uuid.UUID) {
	value := personalInformationID
	c.personalInformationID = &value
}
