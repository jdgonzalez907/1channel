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
	username  string
	createdAt time.Time
}

func NewAgent(id uuid.UUID, username string, createdAt time.Time) (*Agent, error) {
	return &Agent{id, username, createdAt}, nil
}

func (a *Agent) ID() uuid.UUID        { return a.id }
func (a *Agent) Username() string     { return a.username }
func (a *Agent) CreatedAt() time.Time { return a.createdAt }
