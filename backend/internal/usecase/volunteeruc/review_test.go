package volunteeruc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/memory"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/volunteeruc"
)

func TestReviewRejectRequiresPersianReason(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	vid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uuid.New(), Status: domain.StatusPending, FullName: "سارا محمدی",
	})
	_, err := svc.Review(context.Background(), uuid.New(), vid, "reject", "  ")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want invalid input, got %v", err)
	}
	if err.Error() != "برای رد کردن درخواست باید دلیل ثبت شود" {
		t.Fatalf("got %q", err.Error())
	}
}

func TestReviewRejectRecordsHistory(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	vid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uuid.New(), Status: domain.StatusPending, FullName: "سارا محمدی",
	})
	got, err := svc.Review(context.Background(), uuid.New(), vid, "reject", "مدارک ناخوانا است")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.StatusRejected {
		t.Fatalf("status=%s", got.Status)
	}
	if len(got.History) == 0 || got.History[0].Comment != "مدارک ناخوانا است" {
		t.Fatalf("history=%+v", got.History)
	}
}

func TestSetStatusRejectedRequiresReason(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	vid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uuid.New(), Status: domain.StatusApproved,
	})
	_, err := svc.SetStatus(context.Background(), uuid.New(), vid, "rejected", "")
	if err == nil || err.Error() != "برای رد کردن درخواست باید دلیل ثبت شود" {
		t.Fatalf("got %v", err)
	}
}

func TestSetStatusRequiresReason(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	vid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uuid.New(), Status: domain.StatusPending,
	})
	_, err := svc.SetStatus(context.Background(), uuid.New(), vid, "approved", "")
	if err == nil || err.Error() != "برای تغییر وضعیت باید دلیل ثبت شود" {
		t.Fatalf("got %v", err)
	}
}

func TestDeleteDocumentOnlyBeforeAdminStatus(t *testing.T) {
	store := memory.New()
	adapter := memory.VolunteerAdapter{S: store}
	svc := volunteeruc.New(nil, adapter, nil, nil, nil, domain.RealClock{})
	uid, vid, did := uuid.New(), uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uid, Status: domain.StatusDraft,
	})
	_ = adapter.AddDocument(context.Background(), &domain.Document{ID: did, VolunteerID: vid, Kind: domain.DocNationalID, FileName: "id.jpg"})
	if err := svc.DeleteMyDocument(context.Background(), uid, did); err != nil {
		t.Fatal(err)
	}
	_ = adapter.AddDocument(context.Background(), &domain.Document{ID: did, VolunteerID: vid, Kind: domain.DocNationalID, FileName: "id.jpg"})
	v, _ := store.GetVolunteer(context.Background(), vid)
	v.Status = domain.StatusApproved
	_ = store.UpdateVolunteer(context.Background(), v)
	err := svc.DeleteMyDocument(context.Background(), uid, did)
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

func TestAddCommentAppearsInHistory(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	vid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uuid.New(), Status: domain.StatusPending,
	})
	got, err := svc.AddComment(context.Background(), uuid.New(), vid, "لطفا شماره نظام پزشکی را هم بارگذاری کنید")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.History) != 1 || got.History[0].EventType != domain.EventComment {
		t.Fatalf("history=%+v", got.History)
	}
}

func TestAdminUpdateAppliesSkills(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	vid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: vid, UserID: uuid.New(), Status: domain.StatusApproved, FullName: "سارا محمدی",
		City: "تهران", Bio: "توانمند",
	})
	sid := uuid.New()
	ids := []uuid.UUID{sid}
	got, err := svc.AdminUpdate(context.Background(), uuid.New(), vid, volunteeruc.ProfileInput{
		City: "تهران", Bio: "توانمند", SkillIDs: &ids,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Skills) != 1 || got.Skills[0].SkillID != sid {
		t.Fatalf("skills=%+v", got.Skills)
	}
	if got.City != "تهران" || got.Bio != "توانمند" {
		t.Fatalf("profile wiped: city=%q bio=%q", got.City, got.Bio)
	}
}
