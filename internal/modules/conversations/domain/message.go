package domain

import (
	"errors"
	"time"

	"uuid"

	"github.com/rivo/uniseg"
)

const (
	MinTextLength = 1
	MaxTextLength = 1000
)

var (
	ErrMessageInvalidID            = errors.New("message identifier is invalid")
	ErrMessageEmptyText            = errors.New("message text cannot be empty")
	ErrMessageTextTooLong          = errors.New("message text exceeds maximum length")
	ErrMessageExternalIDInvalid    = errors.New("message external identifier is invalid")
	ErrMessageExternalIDAlreadySet = errors.New("message external identifier is already assigned")
	ErrMessageInvalidOwner         = errors.New("message must have exactly one owner (agent or contact)")
	ErrMessageFailed               = errors.New("message failed to deliver, no modifications allowed")
	ErrMessageNotFound             = errors.New("message not found")
	ErrMessageAlreadyDeleted       = errors.New("message is already deleted")
	ErrMessageNotText              = errors.New("message is not a text message")
	ErrMessageNotFromAgent         = errors.New("message was not sent by an agent")
)

type Message struct {
	id         uuid.UUID
	status     MessageStatus
	msgType    MessageType
	text       *string
	agentID    *uuid.UUID
	contactID  *uuid.UUID
	externalID *string
	sentAt     time.Time
	readAt     *time.Time
	editedAt   *time.Time
	deletedAt  *time.Time
}

func NewMessage(
	id uuid.UUID,
	status MessageStatus,
	msgType MessageType,
	text *string,
	agentID *uuid.UUID,
	contactID *uuid.UUID,
	externalID *string,
	sentAt time.Time,
	readAt *time.Time,
	editedAt *time.Time,
	deletedAt *time.Time,
) (*Message, error) {
	if id == uuid.Nil() {
		return nil, ErrMessageInvalidID
	}

	if (agentID == nil && contactID == nil) || (agentID != nil && contactID != nil) {
		return nil, ErrMessageInvalidOwner
	}

	if externalID != nil && *externalID == "" {
		return nil, ErrMessageExternalIDInvalid
	}

	if msgType == MessageTypeText {
		if text == nil {
			return nil, ErrMessageEmptyText
		}

		length := uniseg.GraphemeClusterCount(*text)

		if length < MinTextLength {
			return nil, ErrMessageEmptyText
		}

		if length > MaxTextLength {
			return nil, ErrMessageTextTooLong
		}
	}
	return newMessage(id, status, msgType, text, agentID, contactID, externalID, sentAt, readAt, editedAt, deletedAt), nil
}

func RehydrateMessage(
	id uuid.UUID,
	status MessageStatus,
	msgType MessageType,
	text *string,
	agentID *uuid.UUID,
	contactID *uuid.UUID,
	externalID *string,
	sentAt time.Time,
	readAt *time.Time,
	editedAt *time.Time,
	deletedAt *time.Time,
) *Message {
	return newMessage(id, status, msgType, text, agentID, contactID, externalID, sentAt, readAt, editedAt, deletedAt)
}

func newMessage(
	id uuid.UUID,
	status MessageStatus,
	msgType MessageType,
	text *string,
	agentID *uuid.UUID,
	contactID *uuid.UUID,
	externalID *string,
	sentAt time.Time,
	readAt *time.Time,
	editedAt *time.Time,
	deletedAt *time.Time,
) *Message {
	return &Message{
		id:         id,
		status:     status,
		msgType:    msgType,
		text:       text,
		agentID:    agentID,
		contactID:  contactID,
		externalID: externalID,
		sentAt:     sentAt,
		readAt:     readAt,
		editedAt:   editedAt,
		deletedAt:  deletedAt,
	}
}

func (m *Message) ID() uuid.UUID         { return m.id }
func (m *Message) Status() MessageStatus { return m.status }
func (m *Message) Type() MessageType     { return m.msgType }
func (m *Message) Text() *string         { return m.text }
func (m *Message) AgentID() *uuid.UUID   { return m.agentID }
func (m *Message) ContactID() *uuid.UUID { return m.contactID }
func (m *Message) SentAt() time.Time     { return m.sentAt }
func (m *Message) ReadAt() *time.Time    { return m.readAt }
func (m *Message) EditedAt() *time.Time  { return m.editedAt }
func (m *Message) DeletedAt() *time.Time { return m.deletedAt }
func (m *Message) ExternalID() *string   { return m.externalID }

func (m *Message) AssignExternalID(id string) error {
	if m.externalID != nil {
		return ErrMessageExternalIDAlreadySet
	}

	if id == "" {
		return ErrMessageExternalIDInvalid
	}

	m.externalID = &id
	return nil
}

func (m *Message) MarkAsRead(at time.Time) {
	m.readAt = &at
	m.status = MessageStatusRead
}

func (m *Message) ensureNotFailed() error {
	if m.status == MessageStatusFailed {
		return ErrMessageFailed
	}

	return nil
}

func (m *Message) ensureEditable(at time.Time) error {
	if err := m.ensureNotFailed(); err != nil {
		return err
	}

	if m.deletedAt != nil && !at.Before(*m.deletedAt) {
		return ErrMessageAlreadyDeleted
	}

	return nil
}

func (m *Message) EditText(newText string, at time.Time) error {
	if err := m.ensureEditable(at); err != nil {
		return err
	}

	if m.text == nil {
		return ErrMessageNotText
	}

	if m.editedAt != nil && !at.After(*m.editedAt) {
		return nil
	}

	length := uniseg.GraphemeClusterCount(newText)

	if length < MinTextLength {
		return ErrMessageEmptyText
	}

	if length > MaxTextLength {
		return ErrMessageTextTooLong
	}

	m.text = &newText
	m.editedAt = &at
	return nil
}

func (m *Message) Delete(at time.Time) error {
	if err := m.ensureNotFailed(); err != nil {
		return err
	}

	if m.deletedAt != nil {
		return ErrMessageAlreadyDeleted
	}

	m.deletedAt = &at
	m.status = MessageStatusDeleted
	return nil
}

func (m *Message) MarkAsFailed() error {
	if m.agentID == nil {
		return ErrMessageNotFromAgent
	}

	m.status = MessageStatusFailed
	return nil
}
