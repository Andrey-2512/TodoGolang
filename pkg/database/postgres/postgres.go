package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewClient(DBPath string, maxConns int32, idleConns int32) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig(DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config db: %w", err)
	}
	config.MaxConns = maxConns
	config.MinConns = idleConns

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool db: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil

}
