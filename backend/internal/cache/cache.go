package cache

import (
	"context"
	"errors"
	"time"
)

var ErrMiss = errors.New("cache: key not found")

type Client interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}
