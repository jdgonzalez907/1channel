package domain

import (
	"time"
)

type MessageStatus string

const (
	Sent      MessageStatus = "sent"
	Delivered MessageStatus = "delivered"
	Read      MessageStatus = "read"
)

type MessageType string

const (
	Text  MessageType = "text"
	Audio MessageType = "audio"
	Image MessageType = "image"
	Video MessageType = "video"
	File  MessageType = "file"
)

var (
	AttachmentTypes []MessageType = []MessageType{Audio, Image, Video, File}
)

type Message struct {
	id          string
	externalID  string
	status      MessageStatus
	messageType MessageType
	media       *Media
	content     *string
	agentID     *int64
	createdAt   time.Time
	updatedAt   *time.Time
	readAt      *time.Time
	editedAt    *time.Time
	deletedAt   *time.Time
}

func (m *Message) ID() string               { return m.id }
func (m *Message) ExternalID() string       { return m.externalID }
func (m *Message) Status() MessageStatus    { return m.status }
func (m *Message) MessageType() MessageType { return m.messageType }
func (m *Message) Media() *Media            { return m.media }
func (m *Message) Content() *string         { return m.content }
func (m *Message) AgentID() *int64          { return m.agentID }
func (m *Message) CreatedAt() time.Time     { return m.createdAt }
func (m *Message) UpdatedAt() *time.Time    { return m.updatedAt }
func (m *Message) ReadAt() *time.Time       { return m.readAt }
func (m *Message) EditedAt() *time.Time     { return m.editedAt }
func (m *Message) DeletedAt() *time.Time    { return m.deletedAt }

func NewMessage(
	id, externalID string,
	status MessageStatus,
	messageType MessageType,
	media *Media,
	content *string,
	agentID *int64,
	createdAt time.Time,
	updatedAt, readAt, editedAt, deletedAt *time.Time,
) *Message {
	return &Message{
		id,
		externalID,
		status,
		messageType,
		media,
		content,
		agentID,
		createdAt,
		updatedAt,
		readAt,
		editedAt,
		deletedAt,
	}
}
