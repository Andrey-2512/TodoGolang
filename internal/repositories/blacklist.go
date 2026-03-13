package repositories

import (
	"context"
	"fmt"
	"time"
	"todo/domain/entity"

	"github.com/redis/go-redis/v9"
)

func NewBlacklistRepository(client *redis.Client, blacklistPrefix string) entity.BlacklistRepository {
	return &blacklistRepo{redisClient: client, prefix: blacklistPrefix}
}

type blacklistRepo struct {
	redisClient *redis.Client
	prefix      string
}

func (b *blacklistRepo) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	val, err := b.redisClient.Exists(ctx, b.prefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if jti is blacklisted: %w", err)
	}

	return val > 0, nil
}

func (b *blacklistRepo) Add(ctx context.Context, jti string, exp time.Duration) error {
	cmd := b.redisClient.Set(ctx, b.prefix+jti, "true", exp)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("failed to add to blacklist: %w", err)
	}
	return nil
}
