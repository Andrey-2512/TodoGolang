package repositories

import (
	"context"
	"fmt"
	"time"
	"todo/domain/apperrors"
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

func (b *blacklistRepo) AddNX(ctx context.Context, jti string, exp time.Duration) error {
	res, err := b.redisClient.SetArgs(ctx, b.prefix+jti, "true", redis.SetArgs{
		Mode: "NX",
		TTL:  exp,
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to add to blacklist: %w", err)
	}

	if res != "OK" {
		return apperrors.ErrInvalidToken
	}

	return nil
}
