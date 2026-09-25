package domain

import (
	"errors"
	"time"

	"github.com/rivo/uniseg"
	"uuid"
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
	sentAt time.Time,
	readAt *time.Time,
	editedAt *time.Time,
	deletedAt *time.Time,
) (*Message, error) {
	if id == uuid.Nil() {
		return nil, ErrMessageInvalidID
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
	return &Message{
		id:        id,
		status:    status,
		msgType:   msgType,
		text:      text,
		agentID:   agentID,
		contactID: contactID,
		sentAt:    sentAt,
		readAt:    readAt,
		editedAt:  editedAt,
		deletedAt: deletedAt,
	}, nil
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
