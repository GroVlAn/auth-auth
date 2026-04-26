package database

import (
	"context"
	"errors"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain/e"
	"github.com/redis/go-redis/v9"
)

type Options struct {
	Addr     string
	Password string
	DB       int
}

type RedisClient struct {
	rdb     *redis.Client
	timeout time.Duration
}

func New(opts Options, timeout time.Duration) *RedisClient {
	return &RedisClient{
		rdb: redis.NewClient(&redis.Options{
			Addr:     opts.Addr,
			Password: opts.Password,
			DB:       opts.DB,
		}),
		timeout: timeout,
	}
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.rdb.Set(ctx, key, value, exp).Err()
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", e.ErrRedisNotFound
	}

	if err != nil {
		return "", err
	}

	return val, nil
}

func (c *RedisClient) Del(ctx context.Context, keys ...string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.rdb.Del(ctx, keys...).Err()
}

func (c *RedisClient) SAdd(ctx context.Context, key string, members ...string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := make([]interface{}, len(members))
	for i, v := range members {
		args[i] = v
	}

	return c.rdb.SAdd(ctx, key, args...).Err()
}

func (c *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	vals, err := c.rdb.SMembers(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, e.ErrRedisNotFound
		}

		return nil, err
	}

	return vals, nil
}

func (c *RedisClient) SRem(ctx context.Context, key string, members ...string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	args := make([]interface{}, len(members))

	for i, v := range members {
		args[i] = v
	}

	return c.rdb.SRem(ctx, key, args...).Err()
}

func (c *RedisClient) Shutdown() error {
	return c.rdb.Close()
}
