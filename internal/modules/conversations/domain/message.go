package domain

import (
	"time"
	"uuid"
)

type Message struct {
	id           uuid.UUID
	externalID   *string
	text         string
	status       MessageStatus
	agentID      *uuid.UUID
	contactID    *uuid.UUID
	registeredAt time.Time
	updatedAt    *time.Time
	sentAt       *time.Time
	deliveredAt  *time.Time
	readAt       *time.Time
	failedAt     *time.Time
	deletedAt    *time.Time
}

func NewMessage(
	id uuid.UUID,
	externalID *string,
	text string,
	status MessageStatus,
	agentID *uuid.UUID,
	contactID *uuid.UUID,
	registeredAt time.Time,
	updatedAt *time.Time,
	sentAt *time.Time,
	deliveredAt *time.Time,
	readAt *time.Time,
	failedAt *time.Time,
	deletedAt *time.Time,
) (*Message, error) {
	return &Message{
		id,
		externalID,
		text,
		status,
		agentID,
		contactID,
		registeredAt,
		updatedAt,
		sentAt,
		deliveredAt,
		readAt,
		failedAt,
		deletedAt,
	}, nil
}

func (m *Message) ID() uuid.UUID           { return m.id }
func (m *Message) ExternalID() *string     { return m.externalID }
func (m *Message) Text() string            { return m.text }
func (m *Message) Status() MessageStatus   { return m.status }
func (m *Message) AgentID() *uuid.UUID     { return m.agentID }
func (m *Message) ContactID() *uuid.UUID   { return m.contactID }
func (m *Message) RegisteredAt() time.Time { return m.registeredAt }
func (m *Message) UpdatedAt() *time.Time   { return m.updatedAt }
func (m *Message) SentAt() *time.Time      { return m.sentAt }
func (m *Message) DeliveredAt() *time.Time { return m.deliveredAt }
func (m *Message) ReadAt() *time.Time      { return m.readAt }
func (m *Message) FailedAt() *time.Time    { return m.failedAt }
func (m *Message) DeletedAt() *time.Time   { return m.deletedAt }

func (m *Message) UpdateText(text string, at time.Time) {
	m.text = text
	m.updatedAt = &at
}

func (m *Message) AssignExternalID(externalID string) bool {
	if m.externalID != nil {
		return false
	}

	m.externalID = &externalID

	return true
}

func (m *Message) MarkAsSent(at time.Time) bool {
	return m.markStatusIfForward(Sent, at)
}

func (m *Message) MarkAsDelivered(at time.Time) bool {
	return m.markStatusIfForward(Delivered, at)
}

func (m *Message) MarkAsRead(at time.Time) bool {
	return m.markStatusIfForward(Read, at)
}

func (m *Message) MarkAsFailed(at time.Time) bool {
	return m.markStatusIfForward(Failed, at)
}

func (m *Message) MarkAsDeleted(at time.Time) bool {
	return m.markStatusIfForward(Deleted, at)
}

func (m *Message) markStatusIfForward(newStatus MessageStatus, at time.Time) bool {
	if m.status == Deleted || m.status == Failed {
		return false
	}

	if MessageStatusRank[newStatus] <= MessageStatusRank[m.status] {
		return false
	}

	m.status = newStatus

	switch newStatus {
	case Sent:
		m.sentAt = &at
	case Delivered:
		m.deliveredAt = &at
	case Read:
		m.readAt = &at
	case Failed:
		m.failedAt = &at
	case Deleted:
		m.deletedAt = &at
	}

	return true
}
