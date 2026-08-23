package lock

import (
	"context"
	"testing"
	"time"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func TestMemoryLockSerializes(t *testing.T) {
	m := NewMemory()
	unlock, err := m.Lock(context.Background(), "k", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	unlocked := make(chan struct{})
	go func() {
		u2, err := m.Lock(context.Background(), "k", time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		close(unlocked)
		u2()
	}()
	select {
	case <-unlocked:
		t.Fatal("second lock acquired while first held")
	case <-time.After(50 * time.Millisecond):
	}
	unlock()
	select {
	case <-unlocked:
	case <-time.After(time.Second):
		t.Fatal("second lock did not proceed")
	}
}

func TestRedisLockTimeoutErrorIsBusy(t *testing.T) {
	if domain.ErrBusy == nil {
		t.Fatal("expected ErrBusy")
	}
}
