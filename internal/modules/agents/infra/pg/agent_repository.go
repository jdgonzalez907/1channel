package pg

import (
	"context"
	"errors"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
)

type agentRepository struct {
	db *pgdb.DB
}

func NewAgentRepository(db *pgdb.DB) domain.AgentRepository {
	return &agentRepository{db: db}
}

func (r *agentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	row, err := r.db.Queries.FindAgentByID(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return domain.NewAgent(pgdb.FromUUID(row.ID), pgdb.FromTimestamp(row.CreatedAt))
}
