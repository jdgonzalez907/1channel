package pgdb

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

type Pool interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Close()
}

type DB struct {
	Pool    Pool
	Queries *sqlc.Queries
}

func New(ctx context.Context, dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &DB{Pool: pool, Queries: sqlc.New(pool)}, nil
}

func (db *DB) InTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(db.Queries.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
