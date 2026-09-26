package pg

import (
	"context"
	"errors"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

type conversationRepository struct {
	db *pgdb.DB
}

func NewConversationRepository(db *pgdb.DB) domain.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) FindWithoutMessages(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	row, err := r.db.Queries.FindConversationWithoutMessages(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return toConversation(row, nil)
}

func (r *conversationRepository) FindOpenWithMessageExternalIDsByContactID(ctx context.Context, contactID uuid.UUID) (*domain.Conversation, error) {
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

	messages, err := toMessages(messageRows)
	if err != nil {
		return nil, err
	}

	return toConversation(row, messages)
}

func (r *conversationRepository) FindWithContactUnreadMessagesByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
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

	messages, err := toMessages(messageRows)
	if err != nil {
		return nil, err
	}

	return toConversation(row, messages)
}

func (r *conversationRepository) FindWithMessageByExternalID(ctx context.Context, externalMessageID string) (*domain.Conversation, error) {
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

	message, err := toMessage(messageRow)
	if err != nil {
		return nil, err
	}

	return toConversation(conversationRow, []*domain.Message{message})
}

func (r *conversationRepository) Save(ctx context.Context, conversation *domain.Conversation) error {
	return r.db.InTx(ctx, func(q *sqlc.Queries) error {
		if err := q.UpsertConversation(ctx, sqlc.UpsertConversationParams{
			ID:         pgdb.UUID(conversation.ID()),
			Status:     conversation.Status().String(),
			AgentID:    pgdb.UUIDPtr(conversation.AgentID()),
			ContactID:  pgdb.UUIDPtr(conversation.ContactID()),
			CreatedAt:  pgdb.Timestamp(conversation.CreatedAt()),
			UpdatedAt:  pgdb.TimestampPtr(conversation.UpdatedAt()),
			FinishedAt: pgdb.TimestampPtr(conversation.FinishedAt()),
		}); err != nil {
			return err
		}

		for _, message := range conversation.DirtyMessages() {
			if err := q.UpsertMessage(ctx, toUpsertMessageParams(message, conversation.ID())); err != nil {
				return err
			}
		}

		return nil
	})
}

func toConversation(row sqlc.Conversation, messages []*domain.Message) (*domain.Conversation, error) {
	status, err := domain.NewConversationStatus(row.Status)
	if err != nil {
		return nil, err
	}

	return domain.RehydrateConversation(
		pgdb.FromUUID(row.ID),
		status,
		messages,
		pgdb.FromUUIDPtr(row.AgentID),
		pgdb.FromUUIDPtr(row.ContactID),
		pgdb.FromTimestamp(row.CreatedAt),
		pgdb.FromTimestampPtr(row.UpdatedAt),
		pgdb.FromTimestampPtr(row.FinishedAt),
	), nil
}

func toMessages(rows []sqlc.Message) ([]*domain.Message, error) {
	messages := make([]*domain.Message, 0, len(rows))
	for _, row := range rows {
		message, err := toMessage(row)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func toMessage(row sqlc.Message) (*domain.Message, error) {
	status, err := domain.NewMessageStatus(row.Status)
	if err != nil {
		return nil, err
	}

	messageType, err := domain.NewMessageType(row.Type)
	if err != nil {
		return nil, err
	}

	return domain.RehydrateMessage(
		pgdb.FromUUID(row.ID),
		status,
		messageType,
		row.Text,
		pgdb.FromUUIDPtr(row.AgentID),
		pgdb.FromUUIDPtr(row.ContactID),
		row.ExternalID,
		pgdb.FromTimestamp(row.SentAt),
		pgdb.FromTimestampPtr(row.ReadAt),
		pgdb.FromTimestampPtr(row.EditedAt),
		pgdb.FromTimestampPtr(row.DeletedAt),
	), nil
}

func toUpsertMessageParams(message *domain.Message, conversationID uuid.UUID) sqlc.UpsertMessageParams {
	return sqlc.UpsertMessageParams{
		ID:             pgdb.UUID(message.ID()),
		ConversationID: pgdb.UUID(conversationID),
		Status:         message.Status().String(),
		Type:           message.Type().String(),
		Text:           message.Text(),
		AgentID:        pgdb.UUIDPtr(message.AgentID()),
		ContactID:      pgdb.UUIDPtr(message.ContactID()),
		ExternalID:     message.ExternalID(),
		SentAt:         pgdb.Timestamp(message.SentAt()),
		ReadAt:         pgdb.TimestampPtr(message.ReadAt()),
		EditedAt:       pgdb.TimestampPtr(message.EditedAt()),
		DeletedAt:      pgdb.TimestampPtr(message.DeletedAt()),
	}
}
