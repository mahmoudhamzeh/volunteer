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
	if err.Error() != "نام کامل الزامی است" {
		t.Fatalf("got %q", err.Error())
	}
}
