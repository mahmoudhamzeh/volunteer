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
		Gender: "female", Occupation: "employee",
	})
	got, err := svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{
		FirstName: "علی", LastName: "رضایی", NationalID: "0023456789", Phone: "09350000000",
		Gender: "male", Occupation: "other", OccupationOther: "نویسنده",
		City: "اصفهان",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.FirstName != "سارا" || got.NationalID != "0012345678" || got.Phone != "09121234567" {
		t.Fatalf("identity changed: %+v", got)
	}
	if got.Gender != "female" || got.Occupation != "employee" || got.OccupationOther != "" {
		t.Fatalf("identity extras changed: %+v", got)
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

func TestSubmitRequiresGenderAndOccupation(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	uid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{
		ID: uuid.New(), UserID: uid, Status: domain.StatusDraft,
		FirstName: "سارا", LastName: "محمدی", FullName: "سارا محمدی",
		NationalID: "0012345678", Phone: "09121234567", BirthDate: "1996-05-12",
	})
	_, err := svc.SubmitForReview(context.Background(), uid)
	if err == nil || err.Error() != "جنسیت را انتخاب کنید" {
		t.Fatalf("want gender required, got %v", err)
	}

	_, err = svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{Gender: "female"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.SubmitForReview(context.Background(), uid)
	if err == nil || err.Error() != "شغل را انتخاب کنید" {
		t.Fatalf("want occupation required, got %v", err)
	}

	_, err = svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{Occupation: "other"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.SubmitForReview(context.Background(), uid)
	if err == nil || err.Error() != "در صورت انتخاب «سایر»، شغل خود را بنویسید" {
		t.Fatalf("want other occupation text, got %v", err)
	}

	got, err := svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{Occupation: "other", OccupationOther: "مترجم"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Occupation != "other" || got.OccupationOther != "مترجم" {
		t.Fatalf("occupation other not saved: %+v", got)
	}

	got, err = svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{Occupation: "teacher", OccupationOther: "مترجم"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Occupation != "teacher" || got.OccupationOther != "" {
		t.Fatalf("other text should clear: %+v", got)
	}
}

func TestInvalidGenderRejected(t *testing.T) {
	store := memory.New()
	svc := volunteeruc.New(nil, memory.VolunteerAdapter{S: store}, nil, nil, nil, domain.RealClock{})
	uid := uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{ID: uuid.New(), UserID: uid, Status: domain.StatusDraft})
	_, err := svc.UpsertProfile(context.Background(), uid, volunteeruc.ProfileInput{Gender: "unknown"})
	if err == nil {
		t.Fatal("want invalid gender")
	}
}
