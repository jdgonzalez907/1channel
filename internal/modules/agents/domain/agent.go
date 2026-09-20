package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrAgentNotFound = errors.New("agent not found")
)

type Agent struct {
	id        uuid.UUID
	createdAt time.Time
}

func NewAgent(id uuid.UUID, createdAt time.Time) (*Agent, error) {
	return &Agent{id, createdAt}, nil
}

func (a *Agent) ID() uuid.UUID        { return a.id }
func (a *Agent) CreatedAt() time.Time { return a.createdAt }
