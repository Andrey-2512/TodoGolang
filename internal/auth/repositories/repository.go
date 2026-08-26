package repositories

import (
	"context"
	"fmt"
	"time"
	"todo/domain/apperrors"

	"github.com/redis/go-redis/v9"
)

func NewWhitelistRepository(client *redis.Client, whitelistPrefix string) *WhitelistRepository {
	return &WhitelistRepository{redisClient: client, prefix: whitelistPrefix}
}

type WhitelistRepository struct {
	redisClient *redis.Client
	prefix      string
}

func (w *WhitelistRepository) Del(ctx context.Context, jti string) error {
	const op = "repositories.WhitelistRepository.Del"
	err := w.redisClient.Del(ctx, w.prefix+jti).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
func (w *WhitelistRepository) ConsumeAndAddToken(ctx context.Context, jti, newJti string, exp time.Duration) error {
	const op = "repositories.WhitelistRepository.ConsumeAndAddToken"
	script := redis.NewScript(
		`local exists = redis.call("EXISTS", KEYS[1])
			if exists == 1 then
				redis.call("DEL", KEYS[1])
				redis.call("SET", KEYS[2], ARGV[1], "PX", ARGV[2])
				return 1
			else
				return 0
			end`,
	)
	res, err := script.Run(ctx, w.redisClient, []string{w.prefix + jti, w.prefix + newJti}, "true", exp.Milliseconds()).Int()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res == 0 {
		return fmt.Errorf("%s: %w", op, apperrors.ErrTokenNotInWhitelist)
	}
	return nil
}

func (w *WhitelistRepository) Add(ctx context.Context, jti string, exp time.Duration) error {
	const op = "repositories.WhitelistRepository.Add"
	err := w.redisClient.Set(ctx, w.prefix+jti, "true", exp).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
