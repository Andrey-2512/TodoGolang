package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewClient(connTimeout time.Duration, DBPath string, maxConns int32, idleConns int32, maxConnLifetime time.Duration) (*pgxpool.Pool, error) {
	const op = "postgres.NewClient"
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	config, err := pgxpool.ParseConfig(DBPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	config.MaxConns = maxConns
	config.MinIdleConns = idleConns
	config.MaxConnLifetime = maxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pool, nil

}
