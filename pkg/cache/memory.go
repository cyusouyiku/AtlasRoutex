package cache

import (
"context"
"sync"
"time"
)

type memoryItem struct {
value     []byte
expireAt  time.Time
neverExpires bool
}

type MemoryStore struct {
mu    sync.RWMutex
items map[string]memoryItem
}

func NewMemoryStore() *MemoryStore {
return &MemoryStore{items: make(map[string]memoryItem)}
}

func (m *MemoryStore) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
m.mu.Lock()
defer m.mu.Unlock()
item := memoryItem{value: append([]byte(nil), value...)}
if ttl <= 0 {
item.neverExpires = true
} else {
item.expireAt = time.Now().Add(ttl)
}
m.items[key] = item
return nil
}

func (m *MemoryStore) Get(_ context.Context, key string) ([]byte, error) {
m.mu.RLock()
item, ok := m.items[key]
m.mu.RUnlock()
if !ok {
return nil, ErrCacheMiss
}
if !item.neverExpires && time.Now().After(item.expireAt) {
m.mu.Lock()
delete(m.items, key)
m.mu.Unlock()
return nil, ErrCacheMiss
}
return append([]byte(nil), item.value...), nil
}

func (m *MemoryStore) Delete(_ context.Context, key string) error {
m.mu.Lock()
defer m.mu.Unlock()
delete(m.items, key)
return nil
}

func (m *MemoryStore) Exists(ctx context.Context, key string) (bool, error) {
_, err := m.Get(ctx, key)
if err == ErrCacheMiss {
return false, nil
}
return err == nil, err
}

func (m *MemoryStore) Close() error { return nil }
