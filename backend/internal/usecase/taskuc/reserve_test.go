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

func setupTask(t *testing.T, capacity int) (*taskuc.Service, *memory.Store, uuid.UUID, []uuid.UUID) {
	t.Helper()
	store := memory.New()
	svc := taskuc.New(memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, lock.NewMemory(), nil, domain.RealClock{})
	taskID := uuid.New()
	_ = store.CreateTask(context.Background(), &domain.Task{
		ID:             taskID,
		Title:          "حضور در بخش اطفال",
		Description:    "همراهی",
		Capacity:       capacity,
		HourWeight:     4,
		Status:         domain.TaskOpen,
		StartsAt:       time.Now().Add(time.Hour),
		EndsAt:         time.Now().Add(3 * time.Hour),
		RequiredSkills: []domain.SkillCategory{},
	})
	var users []uuid.UUID
	for i := 0; i < 12; i++ {
		uid := uuid.New()
		vid := uuid.New()
		users = append(users, uid)
		_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
			ID:       vid,
			UserID:   uid,
			Status:   domain.StatusApproved,
			FullName: "V",
		})
	}
	return svc, store, taskID, users
}

func TestApplyDoesNotConsumeCapacity(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 3)

	var wg sync.WaitGroup
	errCh := make(chan error, len(users))
	for _, uid := range users {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			_, err := svc.Accept(context.Background(), id, taskID)
			errCh <- err
		}(uid)
	}
	wg.Wait()
	close(errCh)

	ok := 0
	for err := range errCh {
		if err != nil {
			t.Fatalf("apply error: %v", err)
		}
		ok++
	}
	if ok != 12 {
		t.Fatalf("applied=%d want 12", ok)
	}
	task, _ := store.GetTask(context.Background(), taskID)
	if task.ReservedCount != 0 {
		t.Fatalf("reserved_count=%d want 0", task.ReservedCount)
	}
}

func TestApproveDoesNotExceedCapacity(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 3)
	ctx := context.Background()
	var asgs []*domain.Assignment
	for _, uid := range users {
		a, err := svc.Accept(ctx, uid, taskID)
		if err != nil {
			t.Fatal(err)
		}
		asgs = append(asgs, a)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(asgs))
	for _, a := range asgs {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			_, err := svc.Approve(ctx, id)
			errCh <- err
		}(a.ID)
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
		t.Fatalf("approved=%d want 3 (full=%d)", ok, full)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 3 {
		t.Fatalf("reserved_count=%d", task.ReservedCount)
	}
}

func TestCannotApplyTwice(t *testing.T) {
	svc, _, taskID, users := setupTask(t, 5)
	ctx := context.Background()
	if _, err := svc.Accept(ctx, users[0], taskID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Accept(ctx, users[0], taskID); err != domain.ErrAlreadyAssigned {
		t.Fatalf("want already assigned, got %v", err)
	}
}

func TestVolunteerCancelThenReapply(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 2)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CancelByVolunteer(ctx, users[0], a.ID); err != nil {
		t.Fatal(err)
	}
	again, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != domain.AssignmentRequested {
		t.Fatalf("status=%s", again.Status)
	}
	if _, err := svc.Approve(ctx, again.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CancelByVolunteer(ctx, users[0], again.ID); err != nil {
		t.Fatal(err)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 0 {
		t.Fatalf("reserved_count=%d after cancel reserved", task.ReservedCount)
	}
}

func TestSetStatusStopsNewRequests(t *testing.T) {
	svc, _, taskID, users := setupTask(t, 3)
	ctx := context.Background()
	if _, err := svc.SetStatus(ctx, taskID, domain.TaskInactive); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Accept(ctx, users[0], taskID); err != domain.ErrNotEligible {
		t.Fatalf("want not eligible, got %v", err)
	}
	if _, err := svc.SetStatus(ctx, taskID, domain.TaskClosed); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Accept(ctx, users[1], taskID); err != domain.ErrNotEligible {
		t.Fatalf("want not eligible after close, got %v", err)
	}
}

func TestRemoteDeliveryThenComplete(t *testing.T) {
	svc, store, _, users := setupTask(t, 2)
	ctx := context.Background()
	taskID := uuid.New()
	_ = store.CreateTask(ctx, &domain.Task{
		ID:          taskID,
		Title:       "طراحی پوستر",
		Description: "دورکار",
		Capacity:    2,
		HourWeight:  6,
		Status:      domain.TaskOpen,
		WorkMode:    domain.WorkRemote,
		StartsAt:    time.Now().Add(time.Hour),
		EndsAt:      time.Now().Add(48 * time.Hour),
	})
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ConfirmAttendance(ctx, a.ID); err == nil {
		t.Fatal("remote task should not confirm attendance")
	}
	if _, err := svc.Complete(ctx, a.ID, 5, 5, 5, ""); err == nil {
		t.Fatal("complete before delivery should fail")
	}
	started, err := svc.StartWork(ctx, users[0], a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != domain.AssignmentInProgress {
		t.Fatalf("status=%s want in_progress", started.Status)
	}
	got, err := svc.SubmitDelivery(ctx, users[0], a.ID, taskuc.DeliveryInput{Note: "فایل پوستر آماده است"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.AssignmentSubmitted {
		t.Fatalf("status=%s", got.Status)
	}
	done, err := svc.Complete(ctx, a.ID, 5, 4, 5, "خوب")
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != domain.AssignmentCompleted {
		t.Fatalf("status=%s", done.Status)
	}
}

func TestOnsiteStartThenSubmitDelivery(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 2)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartWork(ctx, users[0], a.ID); err == nil {
		t.Fatal("start before approve should fail")
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	started, err := svc.StartWork(ctx, users[0], a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != domain.AssignmentInProgress {
		t.Fatalf("status=%s want in_progress", started.Status)
	}
	got, err := svc.SubmitDelivery(ctx, users[0], a.ID, taskuc.DeliveryInput{Note: "تست انجام دادم"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.AssignmentSubmitted {
		t.Fatalf("status=%s", got.Status)
	}
	if got.DeliveryNote != "تست انجام دادم" {
		t.Fatalf("note=%q", got.DeliveryNote)
	}
	done, err := svc.Complete(ctx, a.ID, 5, 5, 5, "")
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != domain.AssignmentCompleted {
		t.Fatalf("status=%s", done.Status)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 1 {
		t.Fatalf("reserved_count=%d", task.ReservedCount)
	}
}

func TestCancelInProgressFreesSeat(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 1)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartWork(ctx, users[0], a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CancelByVolunteer(ctx, users[0], a.ID); err != nil {
		t.Fatal(err)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 0 {
		t.Fatalf("reserved_count=%d after cancel in_progress", task.ReservedCount)
	}
}

func TestAdminAssignVolunteer(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 1)
	ctx := context.Background()
	v0, err := store.GetVolunteerByUser(ctx, users[0])
	if err != nil {
		t.Fatal(err)
	}
	a, err := svc.AssignVolunteer(ctx, taskID, v0.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != domain.AssignmentReserved {
		t.Fatalf("status=%s want reserved", a.Status)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 1 {
		t.Fatalf("reserved_count=%d", task.ReservedCount)
	}
	if _, err := svc.AssignVolunteer(ctx, taskID, v0.ID); err != domain.ErrAlreadyAssigned {
		t.Fatalf("want already assigned, got %v", err)
	}
	v1, _ := store.GetVolunteerByUser(ctx, users[1])
	if _, err := svc.AssignVolunteer(ctx, taskID, v1.ID); err != domain.ErrCapacityFull {
		t.Fatalf("want capacity full, got %v", err)
	}
}

func TestAdminAssignPromotesExistingRequest(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 2)
	ctx := context.Background()
	if _, err := svc.Accept(ctx, users[0], taskID); err != nil {
		t.Fatal(err)
	}
	v0, _ := store.GetVolunteerByUser(ctx, users[0])
	a, err := svc.AssignVolunteer(ctx, taskID, v0.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != domain.AssignmentReserved {
		t.Fatalf("status=%s", a.Status)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 1 {
		t.Fatalf("reserved_count=%d", task.ReservedCount)
	}
}
