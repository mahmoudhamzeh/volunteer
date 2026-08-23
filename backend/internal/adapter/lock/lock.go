package lock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
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
	if ttl <= 0 {
		ttl = 8 * time.Second
	}
	lockKey := "lock:" + key
	token := uuid.NewString()
	deadline := time.Now().Add(ttl)
	for {
		ok, err := r.client.SetNX(ctx, lockKey, token, ttl).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			return func() {
				_ = r.client.Del(context.Background(), lockKey).Err()
			}, nil
		}
		if time.Now().After(deadline) {
			return nil, domain.ErrBusy
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(40 * time.Millisecond):
		}
	}
}

var _ domain.Locker = (*Memory)(nil)
var _ domain.Locker = (*Redis)(nil)
