package entity

import (
	"context"
	"time"
)

type BlacklistRepository interface {
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	AddNX(ctx context.Context, jti string, exp time.Duration) error
}
