package cache

import (
"context"
"errors"
"time"
)

var ErrCacheMiss = errors.New("cache miss")

type Store interface {
Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
Get(ctx context.Context, key string) ([]byte, error)
Delete(ctx context.Context, key string) error
Exists(ctx context.Context, key string) (bool, error)
Close() error
}
