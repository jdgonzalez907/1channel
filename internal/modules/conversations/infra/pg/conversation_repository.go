package pg

import (
	"context"
	"errors"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

type postgresConversationRepository struct {
	db *pgdb.DB
}

func NewConversationRepository(db *pgdb.DB) domain.ConversationRepository {
	return &postgresConversationRepository{db: db}
}

func (r *postgresConversationRepository) FindWithoutMessages(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	row, err := r.db.Queries.FindConversationWithoutMessages(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toConversation(row.ID, row.Status, row.UserID, row.ContactID, row.CreatedAt, row.UpdatedAt, row.FinishedAt, nil), nil
}

func (r *postgresConversationRepository) FindOpenWithMessageExternalIDsByContactID(ctx context.Context, contactID uuid.UUID) (*domain.Conversation, error) {
	row, err := r.db.Queries.FindOpenConversationWithMessageExternalIDsByContactID(ctx, pgdb.UUID(contactID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	messageRows, err := r.db.Queries.ListMessagesWithExternalIDByConversation(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return toConversation(row.ID, row.Status, row.UserID, row.ContactID, row.CreatedAt, row.UpdatedAt, row.FinishedAt, toMessages(messageRows)), nil
}

func (r *postgresConversationRepository) FindWithContactUnreadMessagesByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	row, err := r.db.Queries.FindConversationWithContactUnreadMessagesByID(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	messageRows, err := r.db.Queries.ListUnreadContactMessagesByConversation(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return toConversation(row.ID, row.Status, row.UserID, row.ContactID, row.CreatedAt, row.UpdatedAt, row.FinishedAt, toMessages(messageRows)), nil
}

func (r *postgresConversationRepository) FindWithMessageByExternalID(ctx context.Context, externalMessageID string) (*domain.Conversation, error) {
	messageRow, err := r.db.Queries.FindMessageByExternalID(ctx, &externalMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	conversationRow, err := r.db.Queries.FindConversationWithoutMessages(ctx, messageRow.ConversationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toConversation(conversationRow.ID, conversationRow.Status, conversationRow.UserID, conversationRow.ContactID, conversationRow.CreatedAt, conversationRow.UpdatedAt, conversationRow.FinishedAt, []*domain.Message{toMessage(messageRow)}), nil
}

func (r *postgresConversationRepository) Save(ctx context.Context, conversation *domain.Conversation) error {
	dirtyMessages := conversation.DirtyMessages()

	return r.db.InTx(ctx, func(q *sqlc.Queries) error {
		if err := q.UpsertConversation(ctx, sqlc.UpsertConversationParams{
			ID:         pgdb.UUID(conversation.ID()),
			Status:     conversation.Status().String(),
			UserID:     pgdb.UUIDPtr(conversation.AgentID()),
			ContactID:  pgdb.UUIDPtr(conversation.ContactID()),
			CreatedAt:  pgdb.Timestamp(conversation.CreatedAt()),
			UpdatedAt:  pgdb.TimestampPtr(conversation.UpdatedAt()),
			FinishedAt: pgdb.TimestampPtr(conversation.FinishedAt()),
		}); err != nil {
			return err
		}

		for _, message := range dirtyMessages {
			if err := q.UpsertMessage(ctx, toUpsertMessageParams(message, conversation.ID())); err != nil {
				return err
			}
		}

		if len(dirtyMessages) > 0 {
			if err := q.RefreshConversationLastMessage(ctx, pgdb.UUID(conversation.ID())); err != nil {
				return err
			}
		}

		return nil
	})
}

func toConversation(
	id pgtype.UUID,
	status string,
	userID pgtype.UUID,
	contactID pgtype.UUID,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	finishedAt pgtype.Timestamptz,
	messages []*domain.Message,
) *domain.Conversation {
	return domain.RehydrateConversation(
		pgdb.FromUUID(id),
		domain.ConversationStatus(status),
		messages,
		pgdb.FromUUIDPtr(userID),
		pgdb.FromUUIDPtr(contactID),
		pgdb.FromTimestamp(createdAt),
		pgdb.FromTimestampPtr(updatedAt),
		pgdb.FromTimestampPtr(finishedAt),
	)
}

func toMessages(rows []sqlc.Message) []*domain.Message {
	messages := make([]*domain.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, toMessage(row))
	}

	return messages
}

func toMessage(row sqlc.Message) *domain.Message {
	return domain.RehydrateMessage(
		pgdb.FromUUID(row.ID),
		domain.MessageStatus(row.Status),
		domain.MessageType(row.Type),
		row.Text,
		pgdb.FromUUIDPtr(row.UserID),
		pgdb.FromUUIDPtr(row.ContactID),
		row.ExternalID,
		pgdb.FromTimestamp(row.SentAt),
		pgdb.FromTimestampPtr(row.ReadAt),
		pgdb.FromTimestampPtr(row.EditedAt),
		pgdb.FromTimestampPtr(row.DeletedAt),
	)
}

func toUpsertMessageParams(message *domain.Message, conversationID uuid.UUID) sqlc.UpsertMessageParams {
	return sqlc.UpsertMessageParams{
		ID:             pgdb.UUID(message.ID()),
		ConversationID: pgdb.UUID(conversationID),
		Status:         message.Status().String(),
		Type:           message.Type().String(),
		Text:           message.Text(),
		UserID:         pgdb.UUIDPtr(message.AgentID()),
		ContactID:      pgdb.UUIDPtr(message.ContactID()),
		ExternalID:     message.ExternalID(),
		SentAt:         pgdb.Timestamp(message.SentAt()),
		ReadAt:         pgdb.TimestampPtr(message.ReadAt()),
		EditedAt:       pgdb.TimestampPtr(message.EditedAt()),
		DeletedAt:      pgdb.TimestampPtr(message.DeletedAt()),
	}
}
