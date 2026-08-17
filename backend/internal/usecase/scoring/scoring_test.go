package scoring

import (
	"testing"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func TestCompositeScore(t *testing.T) {
	got, err := CompositeScore(5, 4, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got != 4 {
		t.Fatalf("got %v want 4", got)
	}
	if _, err := CompositeScore(0, 5, 5); err == nil {
		t.Fatal("expected invalid rating")
	}
	if _, err := CompositeScore(6, 5, 5); err == nil {
		t.Fatal("expected invalid rating")
	}
}

func TestUpdateVolunteerTotals(t *testing.T) {
	v := domain.Volunteer{}
	UpdateVolunteerTotals(&v, 5, 6)
	if v.CompletedTasks != 1 || v.AverageScore != 5 || v.TotalHours != 6 {
		t.Fatalf("unexpected first update: %+v", v)
	}
	UpdateVolunteerTotals(&v, 3, 2)
	if v.CompletedTasks != 2 {
		t.Fatalf("completed=%d", v.CompletedTasks)
	}
	if v.AverageScore != 4 {
		t.Fatalf("avg=%v want 4", v.AverageScore)
	}
	if v.TotalHours != 8 {
		t.Fatalf("hours=%v", v.TotalHours)
	}
}

func TestEligibleForTask(t *testing.T) {
	v := domain.Volunteer{
		Status:          domain.StatusApproved,
		SkillCategories: []domain.SkillCategory{domain.SkillArtistic},
		AverageScore:    4.5,
		CompletedTasks:  2,
		EducationField:  "گرافیک",
	}
	task := domain.Task{
		RequiredSkills:    []domain.SkillCategory{domain.SkillArtistic},
		MinScore:          4,
		RequiredEducation: "گرافیک",
	}
	if err := EligibleForTask(v, task); err != nil {
		t.Fatal(err)
	}

	v.Status = domain.StatusPending
	if err := EligibleForTask(v, task); err != domain.ErrNotApproved {
		t.Fatalf("want not approved, got %v", err)
	}

	v.Status = domain.StatusApproved
	v.AverageScore = 2
	if err := EligibleForTask(v, task); err != domain.ErrNotEligible {
		t.Fatalf("want not eligible for low score, got %v", err)
	}

	v.AverageScore = 4.5
	v.SkillCategories = []domain.SkillCategory{domain.SkillMedical}
	if err := EligibleForTask(v, task); err != domain.ErrNotEligible {
		t.Fatalf("want not eligible for skill, got %v", err)
	}
}

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to domain.VolunteerStatus
		ok       bool
	}{
		{domain.StatusDraft, domain.StatusPending, true},
		{domain.StatusPending, domain.StatusApproved, true},
		{domain.StatusPending, domain.StatusRejected, true},
		{domain.StatusPending, domain.StatusDraft, true},
		{domain.StatusRejected, domain.StatusPending, true},
		{domain.StatusApproved, domain.StatusSuspended, true},
		{domain.StatusSuspended, domain.StatusApproved, true},
		{domain.StatusApproved, domain.StatusRejected, false},
		{domain.StatusDraft, domain.StatusApproved, false},
	}
	for _, c := range cases {
		got := domain.CanTransition(c.from, c.to)
		if got != c.ok {
			t.Fatalf("%s -> %s got %v want %v", c.from, c.to, got, c.ok)
		}
	}
}
