package pgdb

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func Timestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func TimestampPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}

	return Timestamp(*t)
}

func FromTimestamp(p pgtype.Timestamptz) time.Time {
	return p.Time
}

func FromTimestampPtr(p pgtype.Timestamptz) *time.Time {
	if !p.Valid {
		return nil
	}

	t := p.Time
	return &t
}
