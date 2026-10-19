package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(url string) (*RedisClient, error) {

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	opt.ReadTimeout = 5 * time.Second
	opt.WriteTimeout = 5 * time.Second

	return &RedisClient{
		client: redis.NewClient(opt),
	}, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}

	if err != nil {
		return "", fmt.Errorf("redis error: %w", err)
	}

	return val, nil
}
