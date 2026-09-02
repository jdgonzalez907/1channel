package config

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresConnection(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	ctxConnect, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dbConfig, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		return nil, err
	}

	dbConfig.MaxConns = cfg.DBMaxConns
	dbConfig.MinConns = cfg.DBMinConns
	dbConfig.MaxConnIdleTime = cfg.DBMaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctxConnect, dbConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctxConnect); err != nil {
		defer pool.Close()
		return nil, err
	}

	return pool, nil
}
