package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient(connTimeout time.Duration, addr, password, username string, db, minIdleConns, poolSize int, connMaxLifetime time.Duration) (*redis.Client, error) {
	const op = "redis.NewClient"
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		Username:        username,
		DB:              db,
		PoolSize:        poolSize,
		MinIdleConns:    minIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	})
	status := client.Ping(ctx)

	if err := status.Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return client, nil
}
