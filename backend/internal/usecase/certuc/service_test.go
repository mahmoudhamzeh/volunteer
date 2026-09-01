package certuc_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/memory"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/certuc"
)

type notes struct {
	items []domain.Notification
}

func (n *notes) Notify(_ context.Context, userID uuid.UUID, title, body string) error {
	n.items = append(n.items, domain.Notification{UserID: userID, Title: title, Body: body})
	return nil
}

func setupCert(t *testing.T) (*certuc.Service, *memory.Store, *notes, *domain.Volunteer, *domain.Assignment) {
	t.Helper()
	store := memory.New()
	n := &notes{}
	svc := certuc.New(memory.CertAdapter{S: store}, memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, n, domain.RealClock{}, "http://localhost:3000")
	userID := uuid.New()
	vol := &domain.Volunteer{
		ID:             uuid.New(),
		UserID:         userID,
		FullName:       "علی علی",
		Status:         domain.StatusApproved,
		CompletedTasks: 1,
		TotalHours:     0,
	}
	_ = store.CreateVolunteer(context.Background(), vol)
	taskID := uuid.New()
	_ = store.CreateTask(context.Background(), &domain.Task{
		ID:         taskID,
		Title:      "ورزش",
		HourWeight: 3,
		Status:     domain.TaskOpen,
		StartsAt:   time.Now(),
		EndsAt:     time.Now().Add(time.Hour),
	})
	asg := &domain.Assignment{
		ID:           uuid.New(),
		TaskID:       taskID,
		VolunteerID:  vol.ID,
		Status:       domain.AssignmentCompleted,
		HoursAwarded: 0,
		CreatedAt:    time.Now(),
	}
	_ = memory.TaskAdapter{S: store}.UpdateAssignment(context.Background(), asg)
	return svc, store, n, vol, asg
}

func TestIssueCompletedAssignmentWithZeroHours(t *testing.T) {
	svc, _, n, _, asg := setupCert(t)
	c, err := svc.IssueForAssignment(context.Background(), asg.ID)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if c.Hours != 3 {
		t.Fatalf("hours=%v want 3 from task weight", c.Hours)
	}
	again, err := svc.IssueForAssignment(context.Background(), asg.ID)
	if err != nil {
		t.Fatalf("reissue: %v", err)
	}
	if again.ID != c.ID {
		t.Fatalf("expected existing certificate, got new id")
	}
	if len(n.items) != 1 {
		t.Fatalf("notify=%d want 1", len(n.items))
	}
}

func TestIssueIncompleteAssignmentRejected(t *testing.T) {
	svc, store, _, _, asg := setupCert(t)
	asg.Status = domain.AssignmentSubmitted
	_ = memory.TaskAdapter{S: store}.UpdateAssignment(context.Background(), asg)
	if _, err := svc.IssueForAssignment(context.Background(), asg.ID); err == nil {
		t.Fatal("incomplete assignment should not get a certificate")
	}
}

func TestRequestAndReviewCertificate(t *testing.T) {
	svc, _, n, vol, asg := setupCert(t)
	req, err := svc.Request(context.Background(), vol.UserID, domain.CertTask, &asg.ID)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if req.Status != domain.CertReqPending {
		t.Fatalf("status=%s", req.Status)
	}
	if _, err := svc.Request(context.Background(), vol.UserID, domain.CertTask, &asg.ID); err == nil {
		t.Fatal("duplicate pending request should fail")
	}
	reviewed, err := svc.ReviewRequest(context.Background(), req.ID, "approve", "")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if reviewed.Status != domain.CertReqApproved || reviewed.CertificateID == nil {
		t.Fatalf("approved=%+v", reviewed)
	}
	certs, err := svc.ListByVolunteer(context.Background(), vol.ID)
	if err != nil || len(certs) != 1 {
		t.Fatalf("certs=%v err=%v", certs, err)
	}
	if len(n.items) == 0 {
		t.Fatal("volunteer should be notified")
	}
}

func TestRejectCertificateRequestRequiresReason(t *testing.T) {
	svc, _, _, vol, asg := setupCert(t)
	req, err := svc.Request(context.Background(), vol.UserID, domain.CertTask, &asg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReviewRequest(context.Background(), req.ID, "reject", ""); err == nil {
		t.Fatal("reject without reason should fail")
	}
	reviewed, err := svc.ReviewRequest(context.Background(), req.ID, "reject", "ساعت کافی نیست")
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if reviewed.Status != domain.CertReqRejected {
		t.Fatalf("status=%s", reviewed.Status)
	}
}

func TestAggregatedRequiresCompletedWork(t *testing.T) {
	store := memory.New()
	svc := certuc.New(memory.CertAdapter{S: store}, memory.TaskAdapter{S: store}, memory.VolunteerAdapter{S: store}, nil, domain.RealClock{}, "")
	vol := &domain.Volunteer{ID: uuid.New(), UserID: uuid.New(), Status: domain.StatusApproved}
	_ = store.CreateVolunteer(context.Background(), vol)
	if _, err := svc.IssueAggregated(context.Background(), vol.ID, time.Now().AddDate(-1, 0, 0), time.Now()); err == nil {
		t.Fatal("aggregated without completed work should fail")
	}
	vol.CompletedTasks = 1
	_ = store.UpdateVolunteer(context.Background(), vol)
	if _, err := svc.IssueAggregated(context.Background(), vol.ID, time.Now().AddDate(-1, 0, 0), time.Now()); err != nil {
		t.Fatalf("aggregated: %v", err)
	}
}

func TestOfficialCertificateRequires90HoursAndGoesToPreparing(t *testing.T) {
	svc, store, _, vol, _ := setupCert(t)
	if _, err := svc.Request(context.Background(), vol.UserID, domain.CertOfficial, nil); err == nil {
		t.Fatal("official request under 90 hours should fail")
	}
	vol.TotalHours = 90
	_ = store.UpdateVolunteer(context.Background(), vol)
	req, err := svc.Request(context.Background(), vol.UserID, domain.CertOfficial, nil)
	if err != nil {
		t.Fatalf("official request: %v", err)
	}
	if req.Status != domain.CertReqPreparing {
		t.Fatalf("status=%s want preparing", req.Status)
	}
	if _, err := svc.Request(context.Background(), vol.UserID, domain.CertOfficial, nil); err == nil {
		t.Fatal("duplicate open official request should fail")
	}
	if _, err := svc.ReviewRequest(context.Background(), req.ID, "deliver", ""); err == nil {
		t.Fatal("deliver before issue should fail")
	}
	ready, err := svc.Review(context.Background(), req.ID, certuc.ReviewInput{Action: "approve"})
	if err != nil {
		t.Fatalf("approve official: %v", err)
	}
	if ready.Status != domain.CertReqReady || ready.CertificateID == nil {
		t.Fatalf("ready=%+v", ready)
	}
	done, err := svc.Review(context.Background(), req.ID, certuc.ReviewInput{Action: "deliver", DeliveryMethod: "in_person"})
	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if done.Status != domain.CertReqDelivered || done.DeliveryMethod != "in_person" || done.DeliveredAt == nil {
		t.Fatalf("delivered=%+v", done)
	}
}
