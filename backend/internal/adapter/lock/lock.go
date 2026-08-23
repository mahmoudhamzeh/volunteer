package lock

import (
	"context"
	"strings"
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

// Resilient uses Redis when writable and falls back to process memory
// if Redis is a read-only replica or otherwise refuses writes.
type Resilient struct {
	redis *Redis
	mem   *Memory
}

func NewResilient(client *redis.Client) *Resilient {
	return &Resilient{redis: NewRedis(client), mem: NewMemory()}
}

func (r *Resilient) Lock(ctx context.Context, key string, ttl time.Duration) (func(), error) {
	unlock, err := r.redis.Lock(ctx, key, ttl)
	if err == nil {
		return unlock, nil
	}
	if isReadOnly(err) {
		return r.mem.Lock(ctx, key, ttl)
	}
	return nil, err
}

func isReadOnly(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "readonly") || strings.Contains(s, "read only") || strings.Contains(s, "read-only")
}

var _ domain.Locker = (*Memory)(nil)
var _ domain.Locker = (*Redis)(nil)
var _ domain.Locker = (*Resilient)(nil)
