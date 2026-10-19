// No Operation Cache Client

package cache

import (
	"context"
	"time"
)

type NoopClient struct{}

func NewNoopClient() Client {
	return &NoopClient{}
}

func (n *NoopClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return nil
}

func (n *NoopClient) Get(ctx context.Context, key string) (string, error) {
	return "", ErrMiss
}
