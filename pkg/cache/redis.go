package cache

import (
"context"
"time"

"github.com/redis/go-redis/v9"
)

type RedisStore struct {
client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
return &RedisStore{client: client}
}

func (r *RedisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
v, err := r.client.Get(ctx, key).Bytes()
if err == redis.Nil {
return nil, ErrCacheMiss
}
return v, err
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) Exists(ctx context.Context, key string) (bool, error) {
n, err := r.client.Exists(ctx, key).Result()
return n > 0, err
}

func (r *RedisStore) Close() error { return r.client.Close() }
