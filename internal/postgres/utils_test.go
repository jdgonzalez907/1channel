package postgres

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"uuid"
)

func TestPgUUIDToUUID(t *testing.T) {
	id := uuid.NewV7()
	pg := pgtype.UUID{Bytes: id, Valid: true}

	result := pgUUIDToUUID(pg)

	assert.Equal(t, id, result)
}

func TestPgUUIDToPtr(t *testing.T) {
	t.Run("valid uuid returns pointer", func(t *testing.T) {
		id := uuid.NewV7()
		pg := pgtype.UUID{Bytes: id, Valid: true}

		result := pgUUIDToPtr(pg)

		require.NotNil(t, result)
		assert.Equal(t, id, *result)
	})

	t.Run("invalid uuid returns nil", func(t *testing.T) {
		pg := pgtype.UUID{Valid: false}

		result := pgUUIDToPtr(pg)

		assert.Nil(t, result)
	})
}

func TestPgUUIDFromUUID(t *testing.T) {
	id := uuid.NewV7()

	result := pgUUIDFromUUID(id)

	assert.True(t, result.Valid)
	assert.Equal(t, id, uuid.UUID(result.Bytes))
}

func TestPtrToPgUUID(t *testing.T) {
	t.Run("non-nil pointer returns valid pgtype", func(t *testing.T) {
		id := uuid.NewV7()

		result := ptrToPgUUID(&id)

		assert.True(t, result.Valid)
		assert.Equal(t, id, uuid.UUID(result.Bytes))
	})

	t.Run("nil pointer returns invalid pgtype", func(t *testing.T) {
		result := ptrToPgUUID(nil)

		assert.False(t, result.Valid)
	})
}

func TestToTimestamptz(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	result := toTimestamptz(now)

	assert.True(t, result.Valid)
	assert.Equal(t, now, result.Time)
}

func TestToNillableTimestamptz(t *testing.T) {
	t.Run("non-nil time returns valid pgtype", func(t *testing.T) {
		now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

		result := toNillableTimestamptz(&now)

		assert.True(t, result.Valid)
		assert.Equal(t, now, result.Time)
	})

	t.Run("nil time returns invalid pgtype", func(t *testing.T) {
		result := toNillableTimestamptz(nil)

		assert.False(t, result.Valid)
	})
}

func TestTimePtr(t *testing.T) {
	t.Run("valid timestamptz returns pointer", func(t *testing.T) {
		now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		ts := pgtype.Timestamptz{Time: now, Valid: true}

		result := timePtr(ts)

		require.NotNil(t, result)
		assert.Equal(t, now, *result)
	})

	t.Run("invalid timestamptz returns nil", func(t *testing.T) {
		ts := pgtype.Timestamptz{Valid: false}

		result := timePtr(ts)

		assert.Nil(t, result)
	})
}

func TestTextPtr(t *testing.T) {
	t.Run("valid text returns pointer", func(t *testing.T) {
		text := pgtype.Text{String: "hello", Valid: true}

		result := textPtr(text)

		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("invalid text returns nil", func(t *testing.T) {
		text := pgtype.Text{Valid: false}

		result := textPtr(text)

		assert.Nil(t, result)
	})
}
