package domain

import (
	"context"
	"uuid"
)

type AgentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Agent, error)
}
