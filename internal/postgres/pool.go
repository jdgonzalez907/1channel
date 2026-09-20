package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jdgonzalez907/1channel/internal/config"
)

func NewPool(ctx context.Context, cfg config.Configuration) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.PostgresUsername(),
		cfg.PostgresPassword(),
		cfg.PostgresHost(),
		cfg.PostgresPort(),
		cfg.PostgresDatabase(),
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres config: %w", err)
	}

	poolConfig.MaxConns = cfg.PostgresMaxConns()
	poolConfig.MinConns = cfg.PostgresMinConns()
	poolConfig.MaxConnLifetime = cfg.PostgresMaxConnLifeTime()
	poolConfig.MaxConnIdleTime = cfg.PostgresMaxConnIdleTime()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return pool, nil
}
