package domain

import (
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name      string
		id        uuid.UUID
		createdAt time.Time
	}{
		{
			name:      "creates agent with assigned fields",
			id:        id,
			createdAt: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			agent, err := NewAgent(tt.id, tt.createdAt)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.id, agent.ID())
			assert.Equal(t, tt.createdAt, agent.CreatedAt())
		})
	}
}
