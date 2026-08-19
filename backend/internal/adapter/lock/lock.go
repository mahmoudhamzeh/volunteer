package lock

import (
	"context"
	"sync"
	"time"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Memory struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewMemory() *Memory {
	return &Memory{locks: map[string]*sync.Mutex{}}
}

func (m *Memory) Lock(_ context.Context, key string, _ time.Duration) (func(), error) {
	m.mu.Lock()
	lk, ok := m.locks[key]
	if !ok {
		lk = &sync.Mutex{}
		m.locks[key] = lk
	}
	m.mu.Unlock()
	lk.Lock()
	return func() { lk.Unlock() }, nil
}

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) Lock(ctx context.Context, key string, ttl time.Duration) (func(), error) {
	ok, err := r.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		// fall back to waiting briefly
		deadline := time.Now().Add(ttl)
		for time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
			ok, err = r.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
			if err != nil {
				return nil, err
			}
			if ok {
				break
			}
		}
		if !ok {
			return nil, domain.ErrBusy
		}
	}
	return func() {
		_ = r.client.Del(context.Background(), "lock:"+key).Err()
	}, nil
}

var _ domain.Locker = (*Memory)(nil)
var _ domain.Locker = (*Redis)(nil)
