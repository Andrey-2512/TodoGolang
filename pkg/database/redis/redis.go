package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string, db, minIdleConns, poolSize int, connMaxLifetime time.Duration) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              db,
		PoolSize:        poolSize,
		MinIdleConns:    minIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	})
}
