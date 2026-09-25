package app

import (
	"testing"
	"time"

	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/stretchr/testify/require"
)

func buildConversation(
	t *testing.T,
	id uuid.UUID,
	status domain.ConversationStatus,
	agentID *uuid.UUID,
	contactID *uuid.UUID,
	finishedAt *time.Time,
	messages ...*domain.Message,
) *domain.Conversation {
	t.Helper()

	return domain.RehydrateConversation(id, status, messages, agentID, contactID, time.Now(), nil, finishedAt)
}

func buildContactMessage(t *testing.T, id uuid.UUID, contactID uuid.UUID, externalID *string, readAt *time.Time) *domain.Message {
	t.Helper()

	text := "hello"
	message, err := domain.NewMessage(
		id,
		domain.MessageStatusSent,
		domain.MessageTypeText,
		&text,
		nil,
		&contactID,
		externalID,
		time.Now(),
		readAt,
		nil,
		nil,
	)
	require.NoError(t, err)

	return message
}

func strPtr(s string) *string {
	return &s
}

func buildDeletedContactMessage(t *testing.T, id uuid.UUID, contactID uuid.UUID, externalID string) *domain.Message {
	t.Helper()

	text := "hello"
	deletedAt := time.Now()
	message, err := domain.NewMessage(
		id,
		domain.MessageStatusDeleted,
		domain.MessageTypeText,
		&text,
		nil,
		&contactID,
		&externalID,
		time.Now(),
		nil,
		nil,
		&deletedAt,
	)
	require.NoError(t, err)

	return message
}

func buildAgentMessageWithExternalID(t *testing.T, id uuid.UUID, agentID uuid.UUID, externalID string) *domain.Message {
	t.Helper()

	text := "hello"
	message, err := domain.NewMessage(
		id,
		domain.MessageStatusSent,
		domain.MessageTypeText,
		&text,
		&agentID,
		nil,
		&externalID,
		time.Now(),
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	return message
}

func buildAgentMessage(t *testing.T, id uuid.UUID, agentID uuid.UUID) *domain.Message {
	t.Helper()

	text := "hello"
	message, err := domain.NewMessage(
		id,
		domain.MessageStatusSent,
		domain.MessageTypeText,
		&text,
		&agentID,
		nil,
		nil,
		time.Now(),
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)

	return message
}
