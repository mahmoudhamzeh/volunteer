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

func TestSubmitForReviewPersianValidation(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	uid := uuid.New()
	vid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID:     vid,
		UserID: uid,
		Status: domain.StatusDraft,
	})
	_, err := svc.SubmitForReview(context.Background(), uid)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want invalid input, got %v", err)
	}
	if err.Error() != "نام را وارد کنید" {
		t.Fatalf("got %q", err.Error())
	}
}

func TestUpsertLocksIdentityWhenApproved(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	uid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: uuid.New(), UserID: uid, Status: domain.StatusApproved,
		FirstName: "سارا", LastName: "محمدی", FullName: "سارا محمدی",
		NationalID: "0012345678", Phone: "09121234567", BirthDate: "1996-05-12",
	})
	got, err := svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{
		FirstName: "علی", LastName: "رضایی", NationalID: "0023456789", Phone: "09350000000",
		City: "اصفهان",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.FirstName != "سارا" || got.NationalID != "0012345678" || got.Phone != "09121234567" {
		t.Fatalf("identity changed: %+v", got)
	}
	if got.City != "اصفهان" {
		t.Fatalf("city=%s", got.City)
	}
}

func TestNationalIDMustBeTenDigits(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	uid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{ID: uuid.New(), UserID: uid, Status: domain.StatusDraft})
	_, err := svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{FirstName: "سارا", LastName: "محمدی", NationalID: "123"})
	if err == nil {
		t.Fatal("want national id error")
	}
}
