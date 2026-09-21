package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/postgres/sqlc"
	"uuid"
)

type conversationRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewConversationRepository(pool *pgxpool.Pool) domain.ConversationRepository {
	return &conversationRepository{pool, sqlc.New(pool)}
}

func (r *conversationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	rows, err := r.queries.FindConversationByID(ctx, pgUUIDFromUUID(id))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	conv := rows[0].Conversation
	msgs := make([]sqlc.Message, len(rows))
	for i, row := range rows {
		msgs[i] = row.Message
	}

	return toDomain(conv, msgs)
}

func (r *conversationRepository) FindLastOpenByContactID(ctx context.Context, contactID uuid.UUID) (*domain.Conversation, error) {
	rows, err := r.queries.FindLastOpenConversationByContactID(ctx, pgUUIDFromUUID(contactID))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	conv := rows[0].Conversation
	msgs := make([]sqlc.Message, len(rows))
	for i, row := range rows {
		msgs[i] = row.Message
	}

	return toDomain(conv, msgs)
}

func (r *conversationRepository) FindWithSpecificMessageByExternalID(ctx context.Context, externalMessageID string) (*domain.Conversation, error) {
	row, err := r.queries.FindWithSpecificMessageByExternalID(ctx, pgtype.Text{String: externalMessageID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return toDomain(row.Conversation, []sqlc.Message{row.Message})
}

func (r *conversationRepository) FindWithSpecificMessageByMessageID(ctx context.Context, messageID uuid.UUID) (*domain.Conversation, error) {
	row, err := r.queries.FindWithSpecificMessageByMessageID(ctx, pgUUIDFromUUID(messageID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return toDomain(row.Conversation, []sqlc.Message{row.Message})
}

func (r *conversationRepository) Save(ctx context.Context, conversation *domain.Conversation) error {
	dirtyMessages := conversation.DirtyMessages()

	if len(dirtyMessages) > 0 {
		return r.saveWithMessages(ctx, conversation, dirtyMessages)
	}

	return r.saveConversation(ctx, conversation)
}

func (r *conversationRepository) saveConversation(ctx context.Context, conversation *domain.Conversation) error {
	return r.queries.UpsertConversation(ctx, toUpsertParams(conversation))
}

func (r *conversationRepository) saveWithMessages(ctx context.Context, conversation *domain.Conversation, dirtyMessages []*domain.Message) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if err := qtx.UpsertConversation(ctx, toUpsertParams(conversation)); err != nil {
		return err
	}

	if err := qtx.BatchUpsertMessages(ctx, toBatchUpsertParams(conversation.ID(), dirtyMessages)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func toDomain(conv sqlc.Conversation, msgs []sqlc.Message) (*domain.Conversation, error) {
	messages := make(map[uuid.UUID]*domain.Message, len(msgs))
	for _, row := range msgs {
		m, err := domain.NewMessage(
			pgUUIDToUUID(row.ID),
			textPtr(row.ExternalID),
			row.Text,
			domain.MessageStatus(row.Status),
			pgUUIDToPtr(row.AgentID),
			pgUUIDToPtr(row.ContactID),
			row.CreatedAt.Time,
			timePtr(row.UpdatedAt),
			timePtr(row.DeletedAt),
			timePtr(row.ReadAt),
		)
		if err != nil {
			return nil, err
		}
		messages[m.ID()] = m
	}

	return domain.NewConversation(
		pgUUIDToUUID(conv.ID),
		domain.ConversationStatus(conv.Status),
		messages,
		int8(conv.UnreadCount),
		pgUUIDToPtr(conv.AgentID),
		pgUUIDToUUID(conv.ContactID),
		conv.CreatedAt.Time,
		timePtr(conv.UpdatedAt),
		timePtr(conv.FinishedAt),
	)
}

func toUpsertParams(conversation *domain.Conversation) sqlc.UpsertConversationParams {
	return sqlc.UpsertConversationParams{
		ID:          pgUUIDFromUUID(conversation.ID()),
		Status:      string(conversation.Status()),
		UnreadCount: int16(conversation.UnreadCount()),
		AgentID:     ptrToPgUUID(conversation.AgentID()),
		ContactID:   pgUUIDFromUUID(conversation.ContactID()),
		CreatedAt:   toTimestamptz(conversation.CreatedAt()),
		UpdatedAt:   toNillableTimestamptz(conversation.UpdatedAt()),
		FinishedAt:  toNillableTimestamptz(conversation.FinishedAt()),
	}
}

func toBatchUpsertParams(conversationID uuid.UUID, messages []*domain.Message) sqlc.BatchUpsertMessagesParams {
	n := len(messages)
	params := sqlc.BatchUpsertMessagesParams{
		Ids:             make([]pgtype.UUID, n),
		ConversationIds: make([]pgtype.UUID, n),
		ExternalIds:     make([]string, n),
		Texts:           make([]string, n),
		MessageTypes:    make([]string, n),
		Statuses:        make([]string, n),
		AgentIds:        make([]pgtype.UUID, n),
		ContactIds:      make([]pgtype.UUID, n),
		CreatedAts:      make([]pgtype.Timestamptz, n),
		UpdatedAts:      make([]pgtype.Timestamptz, n),
		DeletedAts:      make([]pgtype.Timestamptz, n),
		ReadAts:         make([]pgtype.Timestamptz, n),
	}

	for i, m := range messages {
		params.Ids[i] = pgUUIDFromUUID(m.ID())
		params.ConversationIds[i] = pgUUIDFromUUID(conversationID)
		if ext := m.ExternalID(); ext != nil {
			params.ExternalIds[i] = *ext
		}
		params.Texts[i] = m.Text()
		params.MessageTypes[i] = "text"
		params.Statuses[i] = string(m.Status())
		params.AgentIds[i] = ptrToPgUUID(m.AgentID())
		params.ContactIds[i] = ptrToPgUUID(m.ContactID())
		params.CreatedAts[i] = toTimestamptz(m.CreatedAt())
		params.UpdatedAts[i] = toNillableTimestamptz(m.UpdatedAt())
		params.DeletedAts[i] = toNillableTimestamptz(m.DeletedAt())
		params.ReadAts[i] = toNillableTimestamptz(m.ReadAt())
	}

	return params
}
