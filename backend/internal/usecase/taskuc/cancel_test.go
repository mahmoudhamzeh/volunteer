package taskuc_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/lock"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/memory"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/taskuc"
)

func TestCancelReleasesCapacity(t *testing.T) {
	store := memory.New()
	svc := taskuc.New(memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, lock.NewMemory(), nil, domain.RealClock{})
	taskID := uuid.New()
	_ = store.CreateTask(context.Background(), &domain.Task{
		ID: taskID, Title: "t", Description: "d", Capacity: 1, HourWeight: 2,
		Status: domain.TaskOpen, StartsAt: time.Now().Add(time.Hour), EndsAt: time.Now().Add(3 * time.Hour),
	})
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{ID: vid, UserID: uid, Status: domain.StatusApproved, FullName: "V"})
	asg, err := svc.Accept(context.Background(), uid, taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CancelMine(context.Background(), uid, asg.ID); err != nil {
		t.Fatal(err)
	}
	task, _ := store.GetTask(context.Background(), taskID)
	if task.ReservedCount != 0 {
		t.Fatalf("reserved=%d", task.ReservedCount)
	}
	if _, err := svc.Accept(context.Background(), uid, taskID); err != nil {
		t.Fatalf("should accept again after cancel: %v", err)
	}
}

func TestListEligiblePaginatesAfterFilter(t *testing.T) {
	store := memory.New()
	svc := taskuc.New(memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, lock.NewMemory(), nil, domain.RealClock{})
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uid, Status: domain.StatusApproved, FullName: "V",
		SkillCategories: []domain.SkillCategory{domain.SkillArtistic},
	})
	for i := 0; i < 5; i++ {
		_ = store.CreateTask(context.Background(), &domain.Task{
			ID: uuid.New(), Title: "a", Description: "d", Capacity: 2, HourWeight: 1,
			Status: domain.TaskOpen, StartsAt: time.Now().Add(time.Hour), EndsAt: time.Now().Add(3 * time.Hour),
			RequiredSkills: []domain.SkillCategory{domain.SkillArtistic},
		})
	}
	_ = store.CreateTask(context.Background(), &domain.Task{
		ID: uuid.New(), Title: "medical", Description: "d", Capacity: 2, HourWeight: 1,
		Status: domain.TaskOpen, StartsAt: time.Now().Add(time.Hour), EndsAt: time.Now().Add(3 * time.Hour),
		RequiredSkills: []domain.SkillCategory{domain.SkillMedical},
	})
	items, total, err := svc.ListEligible(context.Background(), uid, domain.TaskFilter{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Fatalf("total=%d want 5", total)
	}
	if len(items) != 2 {
		t.Fatalf("page=%d want 2", len(items))
	}
}
