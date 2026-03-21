package entity

import (
	"context"
	"time"
)

type WhitelistRepository interface {
	ConsumeAndAddToken(ctx context.Context, jti, newJti string, exp time.Duration) error
	Del(ctx context.Context, jti string) error
	Add(ctx context.Context, jti string, exp time.Duration) error
}
