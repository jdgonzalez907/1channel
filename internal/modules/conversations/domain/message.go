package domain

import (
	"time"
	"uuid"
)

type Message struct {
	id         uuid.UUID
	externalID *string
	text       string
	status     MessageStatus
	agentID    *uuid.UUID
	contactID  *uuid.UUID
	createdAt  time.Time
	updatedAt  *time.Time
	deletedAt  *time.Time
	readAt     *time.Time
}

func NewMessage(
	id uuid.UUID,
	externalID *string,
	text string,
	status MessageStatus,
	agentID *uuid.UUID,
	contactID *uuid.UUID,
	createdAt time.Time,
	updatedAt *time.Time,
	deletedAt *time.Time,
	readAt *time.Time,
) (*Message, error) {
	return &Message{
		id,
		externalID,
		text,
		status,
		agentID,
		contactID,
		createdAt,
		updatedAt,
		deletedAt,
		readAt,
	}, nil
}

func (m *Message) ID() uuid.UUID         { return m.id }
func (m *Message) ExternalID() *string   { return m.externalID }
func (m *Message) Text() string          { return m.text }
func (m *Message) Status() MessageStatus { return m.status }
func (m *Message) AgentID() *uuid.UUID   { return m.agentID }
func (m *Message) ContactID() *uuid.UUID { return m.contactID }
func (m *Message) CreatedAt() time.Time  { return m.createdAt }
func (m *Message) UpdatedAt() *time.Time { return m.updatedAt }
func (m *Message) DeletedAt() *time.Time { return m.deletedAt }
func (m *Message) ReadAt() *time.Time    { return m.readAt }
func (m *Message) Read(at time.Time) bool {
	next := MessageStatusRank[Deleted]
	current := MessageStatusRank[m.status]
	if current >= next {
		return false
	}

	m.status = Read
	m.readAt = &at

	return true
}
func (m *Message) UpdateText(text string, at time.Time) {
	m.text = text
	m.updatedAt = &at
}
func (m *Message) Delete(at time.Time) bool {
	next := MessageStatusRank[Deleted]
	current := MessageStatusRank[m.status]
	if current >= next {
		return false
	}

	m.status = Deleted
	m.deletedAt = &at

	return true
}
