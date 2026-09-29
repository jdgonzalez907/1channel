package domain

import (
	"errors"
	"strings"
	"time"
	"uuid"

	"github.com/rivo/uniseg"
)

const (
	MaxIdentificationNumberLength = 100
	MaxNameLength                 = 100
	MaxPhoneNumberLength          = 100
	MaxEmailLength                = 254
	MaxAddressLength              = 254
)

var (
	ErrPersonalInformationInvalidID = errors.New("personal information identifier is invalid")
	ErrIdentificationNumberEmpty    = errors.New("identification number cannot be empty")
	ErrIdentificationNumberInvalid  = errors.New("identification number must contain only letters and numbers")
	ErrIdentificationNumberTooLong  = errors.New("identification number exceeds maximum length")
	ErrFirstNameTooLong             = errors.New("first name exceeds maximum length")
	ErrLastNameTooLong              = errors.New("last name exceeds maximum length")
	ErrPhoneNumberTooLong           = errors.New("phone number exceeds maximum length")
	ErrEmailTooLong                 = errors.New("email exceeds maximum length")
	ErrAddressTooLong               = errors.New("address exceeds maximum length")
)

type PersonalInformation struct {
	id                   uuid.UUID
	identificationNumber string
	firstName            *string
	lastName             *string
	phoneNumber          *string
	email                *string
	address              *string
	createdAt            time.Time
	updatedAt            time.Time
}

func NewPersonalInformation(
	id uuid.UUID,
	identificationNumber string,
	firstName *string,
	lastName *string,
	phoneNumber *string,
	email *string,
	address *string,
	createdAt time.Time,
	updatedAt time.Time,
) (*PersonalInformation, error) {
	if id == uuid.Nil() {
		return nil, ErrPersonalInformationInvalidID
	}

	identificationNumber = strings.TrimSpace(identificationNumber)
	if err := ValidateIdentificationNumber(identificationNumber); err != nil {
		return nil, err
	}

	normalizedFirstName, err := normalizeOptional(firstName, ErrFirstNameTooLong, MaxNameLength)
	if err != nil {
		return nil, err
	}

	normalizedLastName, err := normalizeOptional(lastName, ErrLastNameTooLong, MaxNameLength)
	if err != nil {
		return nil, err
	}

	normalizedPhoneNumber, err := normalizeOptional(phoneNumber, ErrPhoneNumberTooLong, MaxPhoneNumberLength)
	if err != nil {
		return nil, err
	}

	normalizedEmail, err := normalizeOptional(email, ErrEmailTooLong, MaxEmailLength)
	if err != nil {
		return nil, err
	}

	normalizedAddress, err := normalizeOptional(address, ErrAddressTooLong, MaxAddressLength)
	if err != nil {
		return nil, err
	}

	return &PersonalInformation{
		id:                   id,
		identificationNumber: identificationNumber,
		firstName:            normalizedFirstName,
		lastName:             normalizedLastName,
		phoneNumber:          normalizedPhoneNumber,
		email:                normalizedEmail,
		address:              normalizedAddress,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
	}, nil
}

func RehydratePersonalInformation(
	id uuid.UUID,
	identificationNumber string,
	firstName *string,
	lastName *string,
	phoneNumber *string,
	email *string,
	address *string,
	createdAt time.Time,
	updatedAt time.Time,
) *PersonalInformation {
	return &PersonalInformation{
		id:                   id,
		identificationNumber: identificationNumber,
		firstName:            firstName,
		lastName:             lastName,
		phoneNumber:          phoneNumber,
		email:                email,
		address:              address,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
	}
}

func (p *PersonalInformation) ID() uuid.UUID                { return p.id }
func (p *PersonalInformation) IdentificationNumber() string { return p.identificationNumber }
func (p *PersonalInformation) FirstName() *string           { return p.firstName }
func (p *PersonalInformation) LastName() *string            { return p.lastName }
func (p *PersonalInformation) PhoneNumber() *string         { return p.phoneNumber }
func (p *PersonalInformation) Email() *string               { return p.email }
func (p *PersonalInformation) Address() *string             { return p.address }
func (p *PersonalInformation) CreatedAt() time.Time         { return p.createdAt }
func (p *PersonalInformation) UpdatedAt() time.Time         { return p.updatedAt }

func (p *PersonalInformation) AssignID(id uuid.UUID) {
	p.id = id
}

func (p *PersonalInformation) UpdatePersonalData(
	firstName *string,
	lastName *string,
	phoneNumber *string,
	email *string,
	address *string,
	at time.Time,
) error {
	normalizedFirstName, err := normalizeOptional(firstName, ErrFirstNameTooLong, MaxNameLength)
	if err != nil {
		return err
	}

	normalizedLastName, err := normalizeOptional(lastName, ErrLastNameTooLong, MaxNameLength)
	if err != nil {
		return err
	}

	normalizedPhoneNumber, err := normalizeOptional(phoneNumber, ErrPhoneNumberTooLong, MaxPhoneNumberLength)
	if err != nil {
		return err
	}

	normalizedEmail, err := normalizeOptional(email, ErrEmailTooLong, MaxEmailLength)
	if err != nil {
		return err
	}

	normalizedAddress, err := normalizeOptional(address, ErrAddressTooLong, MaxAddressLength)
	if err != nil {
		return err
	}

	p.firstName = normalizedFirstName
	p.lastName = normalizedLastName
	p.phoneNumber = normalizedPhoneNumber
	p.email = normalizedEmail
	p.address = normalizedAddress
	p.updatedAt = at

	return nil
}

func IsPersonalInformationValidationError(err error) bool {
	validationErrors := []error{
		ErrPersonalInformationInvalidID,
		ErrIdentificationNumberEmpty,
		ErrIdentificationNumberInvalid,
		ErrIdentificationNumberTooLong,
		ErrFirstNameTooLong,
		ErrLastNameTooLong,
		ErrPhoneNumberTooLong,
		ErrEmailTooLong,
		ErrAddressTooLong,
	}

	for _, validationError := range validationErrors {
		if errors.Is(err, validationError) {
			return true
		}
	}

	return false
}

func ValidateIdentificationNumber(identificationNumber string) error {
	length := uniseg.GraphemeClusterCount(identificationNumber)

	if length < 1 {
		return ErrIdentificationNumberEmpty
	}

	if length > MaxIdentificationNumberLength {
		return ErrIdentificationNumberTooLong
	}

	if !isAlphanumeric(identificationNumber) {
		return ErrIdentificationNumberInvalid
	}

	return nil
}

func normalizeOptional(value *string, tooLongErr error, maxLength int) (*string, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	if uniseg.GraphemeClusterCount(trimmed) > maxLength {
		return nil, tooLongErr
	}

	return &trimmed, nil
}

func isAlphanumeric(value string) bool {
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		default:
			return false
		}
	}

	return true
}
