package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain/e"
)

type TokenRepository struct {
	rc        redisClient
	rkBuilder rkBuilder
}

func NewTokenRepository(rc redisClient, rkBuilder rkBuilder) *TokenRepository {
	return &TokenRepository{
		rc:        rc,
		rkBuilder: rkBuilder,
	}
}

func (tr *TokenRepository) BindRefreshToken(ctx context.Context, jti, sessionID string, exp time.Duration) error {
	key := tr.rkBuilder.RefreshKey(jti)

	if err := tr.rc.Set(ctx, key, sessionID, exp); err != nil {
		return e.NewErrInternal(fmt.Errorf("setting new refresh token: %w", err))
	}

	return nil
}

func (tr *TokenRepository) RefreshToken(ctx context.Context, jti string) (string, error) {
	key := tr.rkBuilder.RefreshKey(jti)

	value, err := tr.rc.Get(ctx, key)
	if err != nil {
		if errors.Is(err, e.ErrRedisNotFound) {
			return "", e.NewErrNotFound(
				fmt.Errorf("getting refresh token: %w", err),
				"refresh token not found",
			)
		}

		return "", e.NewErrInternal(fmt.Errorf("getting refresh token: %w", err))
	}

	return value, nil
}

func (tr *TokenRepository) DelRefreshToken(ctx context.Context, jti string) error {
	key := tr.rkBuilder.RefreshKey(jti)

	if err := tr.rc.Del(ctx, key); err != nil {
		return e.NewErrInternal(fmt.Errorf("deleting refresh token: %w", err))
	}

	return nil
}
