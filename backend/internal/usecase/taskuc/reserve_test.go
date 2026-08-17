package taskuc_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/lock"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/memory"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/taskuc"
)

func TestReserveDoesNotExceedCapacity(t *testing.T) {
	store := memory.New()
	svc := taskuc.New(memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, lock.NewMemory(), nil, domain.RealClock{})

	taskID := uuid.New()
	_ = store.CreateTask(context.Background(), &domain.Task{
		ID:           taskID,
		Title:        "حضور در بخش اطفال",
		Description:  "همراهی",
		Capacity:     3,
		HourWeight:   4,
		Status:       domain.TaskOpen,
		StartsAt:     time.Now().Add(time.Hour),
		EndsAt:       time.Now().Add(3 * time.Hour),
		RequiredSkills: []domain.SkillCategory{},
	})

	var volunteers []uuid.UUID
	for i := 0; i < 12; i++ {
		uid := uuid.New()
		vid := uuid.New()
		volunteers = append(volunteers, uid)
		_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
			ID:     vid,
			UserID: uid,
			Status: domain.StatusApproved,
			FullName: "V",
		})
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(volunteers))
	for _, uid := range volunteers {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			_, err := svc.Accept(context.Background(), id, taskID)
			errCh <- err
		}(uid)
	}
	wg.Wait()
	close(errCh)

	ok, full := 0, 0
	for err := range errCh {
		if err == nil {
			ok++
		} else if err == domain.ErrCapacityFull {
			full++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if ok != 3 {
		t.Fatalf("accepted=%d want 3 (full=%d)", ok, full)
	}
	task, _ := store.GetTask(context.Background(), taskID)
	if task.ReservedCount != 3 {
		t.Fatalf("reserved_count=%d", task.ReservedCount)
	}
}
