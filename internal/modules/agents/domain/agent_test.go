package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	id := uuid.NewV7()
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// Arrange
	// (inputs are inline)

	// Act
	agent, err := NewAgent(id, "jane.doe", createdAt)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, agent)
	assert.Equal(t, id, agent.ID())
	assert.Equal(t, "jane.doe", agent.Username())
	assert.Equal(t, createdAt, agent.CreatedAt())
}
