package taskuc_test

import (
	"context"
	"strings"
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
	if _, err := svc.ConfirmAttendance(ctx, a.ID, taskuc.AttendanceInput{}); err == nil {
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

func TestOnsiteAttendanceWithoutStart(t *testing.T) {
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
	if _, err := svc.StartWork(ctx, users[0], a.ID); err == nil {
		t.Fatal("onsite start should fail")
	}
	if _, err := svc.SubmitDelivery(ctx, users[0], a.ID, taskuc.DeliveryInput{Note: "تست انجام دادم"}); err == nil {
		t.Fatal("onsite delivery should fail")
	}
	if _, err := svc.Complete(ctx, a.ID, 5, 5, 5, ""); err == nil {
		t.Fatal("complete before attendance should fail")
	}
	attended, err := svc.ConfirmAttendance(ctx, a.ID, taskuc.AttendanceInput{})
	if err != nil {
		t.Fatal(err)
	}
	if attended.Status != domain.AssignmentAttended {
		t.Fatalf("status=%s want attended", attended.Status)
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

func TestCancelReservedFreesSeat(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 1)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CancelByVolunteer(ctx, users[0], a.ID); err != nil {
		t.Fatal(err)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 0 {
		t.Fatalf("reserved_count=%d after cancel reserved", task.ReservedCount)
	}
}

func TestMarkAbsentFreesSeat(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 1)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	got, err := svc.MarkAbsent(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.AssignmentAbsent {
		t.Fatalf("status=%s", got.Status)
	}
	task, _ := store.GetTask(ctx, taskID)
	if task.ReservedCount != 0 {
		t.Fatalf("reserved=%d", task.ReservedCount)
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

func TestCreateRejectsEndBeforeStart(t *testing.T) {
	svc, _, _, users := setupTask(t, 1)
	_, err := svc.Create(context.Background(), users[0], taskuc.TaskInput{
		Title:       "فعالیت",
		Description: "شرح",
		StartsAt:    time.Now().Add(2 * time.Hour),
		EndsAt:      time.Now().Add(time.Hour),
		Capacity:    1,
		HourWeight:  1,
	})
	if err == nil || !strings.Contains(err.Error(), "تاریخ پایان باید بعد از تاریخ شروع باشد") {
		t.Fatalf("got %v", err)
	}
}

func TestCloseExpiredMarksPastTasksClosed(t *testing.T) {
	svc, store, _, _ := setupTask(t, 1)
	id := uuid.New()
	now := time.Now()
	if err := store.CreateTask(context.Background(), &domain.Task{
		ID:          id,
		Title:       "منقضی",
		Description: "شرح",
		Capacity:    1,
		HourWeight:  1,
		Status:      domain.TaskOpen,
		StartsAt:    now.Add(-3 * time.Hour),
		EndsAt:      now.Add(-time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.CloseExpired(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetTask(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.TaskClosed {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestCreateRequiresTrainingFields(t *testing.T) {
	svc, _, _, users := setupTask(t, 1)
	now := time.Now().Add(time.Hour)
	in := taskuc.TaskInput{
		Title:            "فعالیت آموزشی",
		Description:      "شرح",
		StartsAt:         now,
		EndsAt:           now.Add(2 * time.Hour),
		Capacity:         2,
		HourWeight:       3,
		RequiresTraining: true,
	}
	if _, err := svc.Create(context.Background(), users[0], in); err == nil || !strings.Contains(err.Error(), "نوع آموزش") {
		t.Fatalf("want training kind error, got %v", err)
	}
	in.TrainingKind = domain.TrainingInPerson
	if _, err := svc.Create(context.Background(), users[0], in); err == nil || !strings.Contains(err.Error(), "محل آموزش") {
		t.Fatalf("want training location error, got %v", err)
	}
	in.TrainingLocation = "سالن آموزش"
	if _, err := svc.Create(context.Background(), users[0], in); err == nil || !strings.Contains(err.Error(), "زمان آموزش") {
		t.Fatalf("want training time error, got %v", err)
	}
	at := now.Add(30 * time.Minute)
	in.TrainingAt = &at
	got, err := svc.Create(context.Background(), users[0], in)
	if err != nil {
		t.Fatal(err)
	}
	if !got.RequiresTraining || got.TrainingKind != domain.TrainingInPerson || got.TrainingLocation != "سالن آموزش" || got.TrainingAt == nil {
		t.Fatalf("training fields not saved: %+v", got)
	}
}

func TestRecurringCopiesTrainingToOccurrences(t *testing.T) {
	svc, store, _, users := setupTask(t, 1)
	loc := time.FixedZone("IRST", 3*3600+30*60)
	at := time.Date(2026, 4, 1, 8, 0, 0, 0, loc)
	parent, err := svc.Create(context.Background(), users[0], taskuc.TaskInput{
		Title:            "بازگشایی قلک",
		Description:      "نوبت با آموزش",
		StartsAt:         time.Date(2026, 4, 1, 6, 0, 0, 0, loc),
		EndsAt:           time.Date(2026, 4, 8, 18, 0, 0, 0, loc),
		HourWeight:       4,
		Kind:             domain.TaskRecurring,
		RequiresTraining: true,
		TrainingKind:     domain.TrainingOnline,
		TrainingLocation: "کلاس آنلاین",
		TrainingAt:       &at,
		Slots: []domain.TaskSlot{
			{Weekday: int(time.Wednesday), Capacity: 2, StartTime: "09:00", EndTime: "13:00"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	items, _, err := memory.TaskAdapter{S: store}.List(context.Background(), domain.TaskFilter{SeriesID: parent.ID, Kind: domain.TaskOccurrence, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("no occurrences")
	}
	for _, it := range items {
		if !it.RequiresTraining || it.TrainingKind != domain.TrainingOnline || it.TrainingLocation != "کلاس آنلاین" || it.TrainingAt == nil {
			t.Fatalf("occurrence missing training: %+v", it)
		}
	}
}

type captureNotes struct {
	mu    sync.Mutex
	items []domain.Notification
}

func (n *captureNotes) Notify(_ context.Context, userID uuid.UUID, title, body string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.items = append(n.items, domain.Notification{UserID: userID, Title: title, Body: body, Kind: domain.NotifyNotice})
	return nil
}

func (n *captureNotes) NotifyReminder(_ context.Context, userID uuid.UUID, title, body string, remindAt time.Time) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	at := remindAt
	n.items = append(n.items, domain.Notification{UserID: userID, Title: title, Body: body, Kind: domain.NotifyReminder, RemindAt: &at})
	return nil
}

func (n *captureNotes) snapshot() []domain.Notification {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]domain.Notification, len(n.items))
	copy(out, n.items)
	return out
}

func TestApproveTrainingNotifiesAndReminds(t *testing.T) {
	store := memory.New()
	notes := &captureNotes{}
	svc := taskuc.New(memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, lock.NewMemory(), notes, domain.RealClock{})
	at := time.Now().Add(24 * time.Hour).UTC()
	taskID := uuid.New()
	if err := store.CreateTask(context.Background(), &domain.Task{
		ID:               taskID,
		Title:            "همراهی با آموزش",
		Description:      "شرح",
		Capacity:         2,
		HourWeight:       4,
		Status:           domain.TaskOpen,
		StartsAt:         time.Now().Add(48 * time.Hour),
		EndsAt:           time.Now().Add(50 * time.Hour),
		RequiresTraining: true,
		TrainingKind:     domain.TrainingWorkshop,
		TrainingLocation: "سالن اجتماعات",
		TrainingAt:       &at,
	}); err != nil {
		t.Fatal(err)
	}
	uid := uuid.New()
	vid := uuid.New()
	if err := store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uid, Status: domain.StatusApproved, FullName: "داوطلب",
	}); err != nil {
		t.Fatal(err)
	}
	asg, err := svc.Accept(context.Background(), uid, taskID)
	if err != nil {
		t.Fatal(err)
	}
	gotAsg, err := svc.Approve(context.Background(), asg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAsg.Status != domain.AssignmentTrainingPending {
		t.Fatalf("status=%s want training_pending", gotAsg.Status)
	}
	got := notes.snapshot()
	var approve, reminder *domain.Notification
	for i := range got {
		if strings.Contains(got[i].Body, "آموزش") && got[i].Kind == domain.NotifyNotice && got[i].Title == "فعالیت تایید شد" {
			approve = &got[i]
		}
		if got[i].Kind == domain.NotifyReminder {
			reminder = &got[i]
		}
	}
	if approve == nil {
		t.Fatalf("missing approve notice with training: %+v", got)
	}
	if !strings.Contains(approve.Body, "کارگاه") || !strings.Contains(approve.Body, "سالن اجتماعات") {
		t.Fatalf("approve body=%s", approve.Body)
	}
	if reminder == nil || reminder.RemindAt == nil {
		t.Fatalf("missing reminder: %+v", got)
	}
	if reminder.Title != "یادآوری آموزش" {
		t.Fatalf("reminder title=%s", reminder.Title)
	}
	if !strings.Contains(approve.Body, "شهریور") && !strings.Contains(approve.Body, "ساعت") {
		t.Fatalf("approve body should use jalali time, got %s", approve.Body)
	}
	if !reminder.RemindAt.Equal(at) {
		t.Fatalf("remind_at=%v want %v", reminder.RemindAt, at)
	}
}

func TestConfirmTrainingAddsCourseAndUnlocksActivity(t *testing.T) {
	store := memory.New()
	svc := taskuc.New(memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, lock.NewMemory(), nil, domain.RealClock{})
	at := time.Now().Add(24 * time.Hour).UTC()
	taskID := uuid.New()
	if err := store.CreateTask(context.Background(), &domain.Task{
		ID:               taskID,
		Title:            "فعالیت آموزشی",
		Description:      "شرح",
		Capacity:         2,
		HourWeight:       4,
		Status:           domain.TaskOpen,
		StartsAt:         time.Now().Add(48 * time.Hour),
		EndsAt:           time.Now().Add(50 * time.Hour),
		RequiresTraining: true,
		TrainingKind:     domain.TrainingInPerson,
		TrainingLocation: "محک",
		TrainingAt:       &at,
	}); err != nil {
		t.Fatal(err)
	}
	uid := uuid.New()
	vid := uuid.New()
	if err := store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uid, Status: domain.StatusApproved, FullName: "داوطلب",
	}); err != nil {
		t.Fatal(err)
	}
	asg, err := svc.Accept(context.Background(), uid, taskID)
	if err != nil {
		t.Fatal(err)
	}
	asg, err = svc.Approve(context.Background(), asg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ConfirmAttendance(context.Background(), asg.ID, taskuc.AttendanceInput{}); err == nil {
		t.Fatal("attendance before training confirm should fail")
	}
	got, err := svc.ConfirmTraining(context.Background(), asg.ID, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.AssignmentReserved {
		t.Fatalf("status=%s want reserved", got.Status)
	}
	courses, err := svc.MyTrainings(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(courses) != 1 {
		t.Fatalf("courses=%d want 1", len(courses))
	}
	if courses[0].TrainingLocation != "محک" {
		t.Fatalf("location=%s", courses[0].TrainingLocation)
	}
}

func TestApproveSkipsTrainingWhenCourseAlreadyOnFile(t *testing.T) {
	store := memory.New()
	svc := taskuc.New(memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, lock.NewMemory(), nil, domain.RealClock{})
	at := time.Now().Add(24 * time.Hour).UTC()
	mk := func(title string) uuid.UUID {
		id := uuid.New()
		if err := store.CreateTask(context.Background(), &domain.Task{
			ID:               id,
			Title:            title,
			Description:      "شرح",
			Capacity:         2,
			HourWeight:       4,
			Status:           domain.TaskOpen,
			StartsAt:         time.Now().Add(48 * time.Hour),
			EndsAt:           time.Now().Add(50 * time.Hour),
			RequiresTraining: true,
			TrainingKind:     domain.TrainingInPerson,
			TrainingLocation: "محک",
			TrainingAt:       &at,
		}); err != nil {
			t.Fatal(err)
		}
		return id
	}
	first := mk("اول")
	second := mk("دوم")
	uid := uuid.New()
	vid := uuid.New()
	if err := store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uid, Status: domain.StatusApproved, FullName: "داوطلب",
	}); err != nil {
		t.Fatal(err)
	}
	a1, err := svc.Accept(context.Background(), uid, first)
	if err != nil {
		t.Fatal(err)
	}
	a1, err = svc.Approve(context.Background(), a1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ConfirmTraining(context.Background(), a1.ID, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	a2, err := svc.Accept(context.Background(), uid, second)
	if err != nil {
		t.Fatal(err)
	}
	a2, err = svc.Approve(context.Background(), a2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a2.Status != domain.AssignmentReserved {
		t.Fatalf("second status=%s want reserved", a2.Status)
	}
}

func TestCreateRecurringExpandsWeekdays(t *testing.T) {
	svc, store, _, users := setupTask(t, 1)
	loc := time.FixedZone("IRST", 3*3600+30*60)
	parent, err := svc.Create(context.Background(), users[0], taskuc.TaskInput{
		Title:       "بازگشایی قلک",
		Description: "نوبت‌های دوشنبه و سه‌شنبه",
		StartsAt:    time.Date(2026, 4, 1, 6, 0, 0, 0, loc),
		EndsAt:      time.Date(2026, 4, 15, 18, 0, 0, 0, loc),
		HourWeight:  4,
		Kind:        domain.TaskRecurring,
		Slots: []domain.TaskSlot{
			{Weekday: int(time.Monday), Capacity: 3, StartTime: "09:00", EndTime: "13:00"},
			{Weekday: int(time.Tuesday), Capacity: 8, StartTime: "10:00", EndTime: "14:00"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if parent.Kind != domain.TaskRecurring {
		t.Fatalf("kind=%s", parent.Kind)
	}
	items, _, err := memory.TaskAdapter{S: store}.List(context.Background(), domain.TaskFilter{SeriesID: parent.ID, Kind: domain.TaskOccurrence, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("no occurrences")
	}
	for _, it := range items {
		wd := time.Weekday(it.Weekday)
		if wd != time.Monday && wd != time.Tuesday {
			t.Fatalf("weekday=%s", wd)
		}
		if it.Capacity != 3 && it.Capacity != 8 {
			t.Fatalf("capacity=%d", it.Capacity)
		}
	}
}

func TestUpdateRecurringReplacesWeekdays(t *testing.T) {
	svc, store, _, users := setupTask(t, 1)
	loc := time.FixedZone("IRST", 3*3600+30*60)
	in := taskuc.TaskInput{
		Title:       "بازگشایی قلک",
		Description: "نوبت‌های یکشنبه و سه‌شنبه",
		StartsAt:    time.Date(2026, 4, 1, 6, 0, 0, 0, loc),
		EndsAt:      time.Date(2026, 4, 15, 18, 0, 0, 0, loc),
		HourWeight:  4,
		Kind:        domain.TaskRecurring,
		Slots: []domain.TaskSlot{
			{Weekday: int(time.Sunday), Capacity: 2, StartTime: "09:00", EndTime: "13:00"},
			{Weekday: int(time.Tuesday), Capacity: 5, StartTime: "10:00", EndTime: "14:00"},
		},
	}
	parent, err := svc.Create(context.Background(), users[0], in)
	if err != nil {
		t.Fatal(err)
	}
	in.Description = "نوبت‌های شنبه و سه‌شنبه"
	in.Slots = []domain.TaskSlot{
		{Weekday: int(time.Saturday), Capacity: 4, StartTime: "09:00", EndTime: "13:00"},
		{Weekday: int(time.Tuesday), Capacity: 5, StartTime: "10:00", EndTime: "14:00"},
	}
	if _, err := svc.Update(context.Background(), parent.ID, in); err != nil {
		t.Fatal(err)
	}
	items, _, err := memory.TaskAdapter{S: store}.List(context.Background(), domain.TaskFilter{SeriesID: parent.ID, Kind: domain.TaskOccurrence, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	var open []domain.Task
	for _, it := range items {
		if it.Status == domain.TaskClosed {
			continue
		}
		open = append(open, it)
		wd := time.Weekday(it.Weekday)
		if wd != time.Saturday && wd != time.Tuesday {
			t.Fatalf("unexpected weekday %s", wd)
		}
	}
	if len(open) == 0 {
		t.Fatal("expected occurrences after edit")
	}
	hasSat, hasSun := false, false
	for _, it := range open {
		if time.Weekday(it.Weekday) == time.Saturday {
			hasSat = true
		}
		if time.Weekday(it.Weekday) == time.Sunday {
			hasSun = true
		}
	}
	if !hasSat {
		t.Fatal("expected Saturday occurrences after edit")
	}
	if hasSun {
		t.Fatal("Sunday occurrences should be removed after edit")
	}
}

func TestVolunteerRatesWithComment(t *testing.T) {
	svc, _, taskID, users := setupTask(t, 1)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ConfirmAttendance(ctx, a.ID, taskuc.AttendanceInput{}); err != nil {
		t.Fatal(err)
	}
	got, err := svc.RateByVolunteer(ctx, users[0], a.ID, 4, "  هماهنگی خوب بود  ")
	if err != nil {
		t.Fatal(err)
	}
	if got.VolunteerRating == nil || *got.VolunteerRating != 4 {
		t.Fatalf("rating=%v", got.VolunteerRating)
	}
	if got.VolunteerComment != "هماهنگی خوب بود" {
		t.Fatalf("comment=%q", got.VolunteerComment)
	}
}

func TestRemoteRevisionThenResubmitThenComplete(t *testing.T) {
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
	if _, err := svc.StartWork(ctx, users[0], a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SubmitDelivery(ctx, users[0], a.ID, taskuc.DeliveryInput{Note: "نسخه اول"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RequestRevision(ctx, a.ID, ""); err == nil {
		t.Fatal("revision without comment should fail")
	}
	rev, err := svc.RequestRevision(ctx, a.ID, "لطفا فایل نهایی را هم بفرستید")
	if err != nil {
		t.Fatal(err)
	}
	if rev.Status != domain.AssignmentRevisionRequested {
		t.Fatalf("status=%s", rev.Status)
	}
	if !rev.Status.OccupiesSeat() || !rev.Status.BlocksReapply() {
		t.Fatal("revision should keep the seat")
	}
	if _, err := svc.Complete(ctx, a.ID, 5, 5, 5, ""); err == nil {
		t.Fatal("complete during revision should fail")
	}
	got, err := svc.SubmitDelivery(ctx, users[0], a.ID, taskuc.DeliveryInput{Note: "نسخه اصلاح‌شده"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.AssignmentSubmitted {
		t.Fatalf("status=%s", got.Status)
	}
	if len(got.History) != 3 {
		t.Fatalf("history=%d want 3: %+v", len(got.History), got.History)
	}
	if got.History[0].Kind != domain.AssignmentEventDelivery || got.History[0].Note != "نسخه اول" {
		t.Fatalf("first event=%+v", got.History[0])
	}
	if got.History[1].Kind != domain.AssignmentEventRevision {
		t.Fatalf("second event=%+v", got.History[1])
	}
	if got.History[2].Kind != domain.AssignmentEventDelivery || got.History[2].Note != "نسخه اصلاح‌شده" {
		t.Fatalf("third event=%+v", got.History[2])
	}
	done, err := svc.Complete(ctx, a.ID, 5, 5, 5, "قبول")
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != domain.AssignmentCompleted {
		t.Fatalf("status=%s", done.Status)
	}
}

func TestRequestRevisionOnsiteFails(t *testing.T) {
	svc, _, taskID, users := setupTask(t, 1)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RequestRevision(ctx, a.ID, "اصلاح کنید"); err == nil {
		t.Fatal("onsite revision should fail")
	}
}

func TestOnsiteAttendanceManualTimes(t *testing.T) {
	svc, _, taskID, users := setupTask(t, 1)
	ctx := context.Background()
	a, err := svc.Accept(ctx, users[0], taskID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	inAt := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	outAt := time.Date(2026, 8, 24, 8, 0, 0, 0, time.UTC)
	if _, err := svc.ConfirmAttendance(ctx, a.ID, taskuc.AttendanceInput{CheckInAt: &inAt, CheckOutAt: &outAt}); err == nil {
		t.Fatal("checkout before checkin should fail")
	}
	outAt = time.Date(2026, 8, 24, 13, 0, 0, 0, time.UTC)
	got, err := svc.ConfirmAttendance(ctx, a.ID, taskuc.AttendanceInput{CheckInAt: &inAt, CheckOutAt: &outAt})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.AssignmentAttended {
		t.Fatalf("status=%s", got.Status)
	}
	if got.CheckInAt == nil || !got.CheckInAt.Equal(inAt) {
		t.Fatalf("check_in=%v", got.CheckInAt)
	}
	if got.CheckOutAt == nil || !got.CheckOutAt.Equal(outAt) {
		t.Fatalf("check_out=%v", got.CheckOutAt)
	}
	laterOut := time.Date(2026, 8, 24, 14, 30, 0, 0, time.UTC)
	updated, err := svc.ConfirmAttendance(ctx, a.ID, taskuc.AttendanceInput{CheckOutAt: &laterOut})
	if err != nil {
		t.Fatal(err)
	}
	if updated.CheckInAt == nil || !updated.CheckInAt.Equal(inAt) {
		t.Fatalf("check_in should stay %v got %v", inAt, updated.CheckInAt)
	}
	if updated.CheckOutAt == nil || !updated.CheckOutAt.Equal(laterOut) {
		t.Fatalf("check_out=%v", updated.CheckOutAt)
	}
}

func TestSuspendedVolunteerCannotSeeOrApplyTasks(t *testing.T) {
	svc, store, taskID, users := setupTask(t, 3)
	uid := users[0]
	v, _ := store.GetVolunteerByUser(context.Background(), uid)
	v.Status = domain.StatusSuspended
	_ = store.UpdateVolunteer(context.Background(), v)

	items, total, err := svc.ListEligible(context.Background(), uid, domain.TaskFilter{Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("suspended volunteer saw tasks: n=%d total=%d", len(items), total)
	}
	_, err = svc.Accept(context.Background(), uid, taskID)
	if err == nil || !strings.Contains(err.Error(), "امکان درخواست") {
		t.Fatalf("want blocked apply, got %v", err)
	}
}
