package pgdb

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureUTCScansTimestamptzInUTC(t *testing.T) {
	m := pgtype.NewMap()
	configureUTC(m)

	instant := time.Date(2026, 9, 27, 15, 4, 5, 0, time.FixedZone("UTC-5", -5*60*60))

	encoded, err := m.Encode(
		pgtype.TimestamptzOID,
		pgtype.BinaryFormatCode,
		pgtype.Timestamptz{Time: instant, Valid: true},
		nil,
	)
	require.NoError(t, err)

	var got pgtype.Timestamptz
	require.NoError(t, m.Scan(pgtype.TimestamptzOID, pgtype.BinaryFormatCode, encoded, &got))

	require.True(t, got.Valid)
	assert.True(t, instant.Equal(got.Time), "el instante debe preservarse")
	assert.Equal(t, time.UTC, got.Time.Location(), "la lectura debe exponerse en UTC")
}
