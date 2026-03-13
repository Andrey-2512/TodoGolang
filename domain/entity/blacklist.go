package entity

import (
	"context"
	"time"
)

type BlacklistRepository interface {
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	Add(ctx context.Context, jti string, exp time.Duration) error
}
