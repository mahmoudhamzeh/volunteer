package httpserver

import (
	"testing"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func TestVolunteerSelfDTOHidesSuspension(t *testing.T) {
	v := &domain.Volunteer{
		Status:          domain.StatusSuspended,
		RejectionReason: "بررسی داخلی",
		History: []domain.VolunteerEvent{
			{EventType: domain.EventApproved, FromStatus: domain.StatusPending, ToStatus: domain.StatusApproved},
			{EventType: domain.EventSuspended, FromStatus: domain.StatusApproved, ToStatus: domain.StatusSuspended, Comment: "secret"},
		},
	}
	d := volunteerSelfDTO(v)
	if d["status"] != domain.StatusApproved {
		t.Fatalf("status=%v", d["status"])
	}
	if d["rejection_reason"] != "" {
		t.Fatalf("reason leaked: %v", d["rejection_reason"])
	}
	hist, _ := d["history"].([]domain.VolunteerEvent)
	if len(hist) != 1 || hist[0].EventType != domain.EventApproved {
		t.Fatalf("history=%+v", hist)
	}
}
