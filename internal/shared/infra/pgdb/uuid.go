package pgdb

import (
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
)

func UUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func UUIDPtr(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{}
	}

	return UUID(*u)
}

func FromUUID(p pgtype.UUID) uuid.UUID {
	return p.Bytes
}

func FromUUIDPtr(p pgtype.UUID) *uuid.UUID {
	if !p.Valid {
		return nil
	}

	u := uuid.UUID(p.Bytes)
	return &u
}
