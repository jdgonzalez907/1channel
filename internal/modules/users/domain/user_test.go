package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()

	tests := []struct {
		name      string
		id        uuid.UUID
		createdAt time.Time
		wantErr   error
	}{
		{
			name:      "creates user with assigned fields",
			id:        id,
			createdAt: now,
		},
		{
			name:      "fails with nil id",
			id:        uuid.Nil(),
			createdAt: now,
			wantErr:   ErrUserInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			user, err := NewUser(tt.id, tt.createdAt)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.id, user.ID())
			assert.Equal(t, tt.createdAt, user.CreatedAt())
		})
	}
}

func TestRehydrateUser(t *testing.T) {
	id := uuid.NewV7()
	now := time.Now()

	// Act
	user := RehydrateUser(id, now)

	// Assert
	assert.Equal(t, id, user.ID())
	assert.Equal(t, now, user.CreatedAt())
}
