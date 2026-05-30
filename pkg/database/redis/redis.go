package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(connTimeout time.Duration, addr, password string, db, minIdleConns, poolSize int, connMaxLifetime time.Duration) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              db,
		PoolSize:        poolSize,
		MinIdleConns:    minIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	})
	status := client.Ping(ctx)

	if err := status.Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return client, nil
}
