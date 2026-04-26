package repository

import (
	"context"
	"time"
)

type redisClient interface {
	Set(ctx context.Context, key string, value interface{}, exp time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error

	SAdd(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...string) error
}

type rkBuilder interface {
	SessionKey(sessionID string) string
	RefreshKey(jti string) string
	UserSessionsKey(userID string) string
	BlacklistKey(jti string) string
}
