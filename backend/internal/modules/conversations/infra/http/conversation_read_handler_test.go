package http

import (
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

func TestStatusesFromTab(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   []string
		ok     bool
	}{
		{name: "open", status: "open", want: []string{"pending", "assigned"}, ok: true},
		{name: "finished", status: "finished", want: []string{"resolved", "expired"}, ok: true},
		{name: "all is rejected", status: "all", want: nil, ok: false},
		{name: "empty is rejected", status: "", want: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := statusesFromTab(tt.status)

			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsConversationVisible(t *testing.T) {
	agentID := uuid.NewV7()
	otherAgentID := uuid.NewV7()

	tests := []struct {
		name         string
		conversation sqlc.FindConversationWithContactByIDRow
		want         bool
	}{
		{name: "pending without agent is visible", conversation: sqlc.FindConversationWithContactByIDRow{Status: "pending"}, want: true},
		{name: "expired without agent is visible", conversation: sqlc.FindConversationWithContactByIDRow{Status: "expired"}, want: true},
		{name: "own is visible", conversation: sqlc.FindConversationWithContactByIDRow{Status: "assigned", UserID: pgdb.UUID(agentID)}, want: true},
		{name: "own expired is visible", conversation: sqlc.FindConversationWithContactByIDRow{Status: "expired", UserID: pgdb.UUID(agentID)}, want: true},
		{name: "other agent is not visible", conversation: sqlc.FindConversationWithContactByIDRow{Status: "assigned", UserID: pgdb.UUID(otherAgentID)}, want: false},
		{name: "expired with other agent is not visible", conversation: sqlc.FindConversationWithContactByIDRow{Status: "expired", UserID: pgdb.UUID(otherAgentID)}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isConversationVisible(tt.conversation, agentID))
		})
	}
}

func TestConversationListItemFromRow(t *testing.T) {
	now := time.Now().UTC()
	conversationID := uuid.NewV7()
	contactID := uuid.NewV7()
	text := "hola"

	row := sqlc.ListConversationsForAgentRow{
		ID:                pgdb.UUID(conversationID),
		Status:            "pending",
		LastMessageAt:     pgdb.Timestamp(now),
		UnreadCount:       3,
		ContactID:         pgdb.UUID(contactID),
		ExternalContactID: "ext-1",
		LastMessageStatus: "sent",
		LastMessageText:   &text,
		LastMessageOwner:  "contact",
	}

	got := conversationListItemFromRow(row)
	assert.Equal(t, conversationID.String(), got.ID)
	assert.Equal(t, "ext-1", got.Contact.ExternalID)
	assert.EqualValues(t, 3, got.UnreadCount)
	assert.Equal(t, "contact", got.LastMessage.Owner)
	require.NotNil(t, got.LastMessage.Text)
	assert.Equal(t, "hola", *got.LastMessage.Text)

	row.LastMessageStatus = "deleted"
	assert.Nil(t, conversationListItemFromRow(row).LastMessage.Text)
}

func TestConversationMessageFromRow(t *testing.T) {
	now := time.Now().UTC()
	messageID := uuid.NewV7()
	agentID := uuid.NewV7()
	contactID := uuid.NewV7()
	text := "hola"

	contactRow := sqlc.ListConversationMessagesPageRow{
		ID:        pgdb.UUID(messageID),
		Status:    "sent",
		Type:      "text",
		Text:      &text,
		ContactID: pgdb.UUID(contactID),
		SentAt:    pgdb.Timestamp(now),
	}

	got := conversationMessageFromRow(contactRow)
	assert.Equal(t, "contact", got.Owner)
	require.NotNil(t, got.Text)
	assert.Equal(t, "hola", *got.Text)
	assert.Nil(t, got.ReadAt)

	agentRow := contactRow
	agentRow.UserID = pgdb.UUID(agentID)
	agentRow.ContactID = pgtype.UUID{}
	assert.Equal(t, "agent", conversationMessageFromRow(agentRow).Owner)

	deletedRow := contactRow
	deletedRow.Status = "deleted"
	deleted := conversationMessageFromRow(deletedRow)
	assert.Nil(t, deleted.Text)
	assert.Equal(t, "deleted", deleted.Status)
}
