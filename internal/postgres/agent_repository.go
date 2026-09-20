package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jdgonzalez907/1channel/internal/modules/agents/domain"
	"github.com/jdgonzalez907/1channel/internal/postgres/sqlc"
	"uuid"
)

type agentRepository struct {
	queries *sqlc.Queries
}

func NewAgentRepository(pool *pgxpool.Pool) domain.AgentRepository {
	return &agentRepository{sqlc.New(pool)}
}

func (r *agentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	row, err := r.queries.FindAgentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return domain.NewAgent(row.Agent.ID, row.Agent.CreatedAt.Time)
}
