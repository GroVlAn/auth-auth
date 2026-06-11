package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain/e"
	"github.com/redis/go-redis/v9"
)

type BlacklistRepository struct {
	rc        *redis.Client
	rkBuilder rkBuilder
}

func NewBlacklistRepository(rc *redis.Client, rkBuilder rkBuilder) *BlacklistRepository {
	return &BlacklistRepository{
		rc:        rc,
		rkBuilder: rkBuilder,
	}
}

func (br *BlacklistRepository) AddToBlackList(ctx context.Context, jti string, exp time.Duration) error {
	key := br.rkBuilder.BlacklistKey(jti)

	if err := br.rc.Set(ctx, key, "1", exp).Err(); err != nil {
		return e.NewErrInternal(fmt.Errorf("adding token to black list: %w", err))

	}

	return nil
}

func (br *BlacklistRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := br.rkBuilder.BlacklistKey(jti)

	val, err := br.rc.Get(ctx, key).Result()

	if err != nil {
		if errors.Is(err, e.ErrRedisNotFound) {
			return false, nil
		}

		return false, e.NewErrInternal(fmt.Errorf("adding token to black list: %w", err))
	}

	return val != "", nil
}
