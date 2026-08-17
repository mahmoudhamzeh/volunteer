package scoring

import "github.com/mahmoudhamzeh/volunteer/backend/internal/domain"

const (
	MinRating = 1
	MaxRating = 5
)

// CompositeScore averages the three admin evaluation dimensions.
func CompositeScore(discipline, expertise, ethics int) (float64, error) {
	for _, v := range []int{discipline, expertise, ethics} {
		if v < MinRating || v > MaxRating {
			return 0, domain.ErrInvalidInput
		}
	}
	return float64(discipline+expertise+ethics) / 3.0, nil
}

// UpdateVolunteerTotals recalculates running average after a completed assignment.
func UpdateVolunteerTotals(v *domain.Volunteer, assignmentScore, hours float64) {
	completed := v.CompletedTasks
	if completed < 0 {
		completed = 0
	}
	total := float64(completed)*v.AverageScore + assignmentScore
	v.CompletedTasks = completed + 1
	v.AverageScore = total / float64(v.CompletedTasks)
	v.TotalHours += hours
}

// EligibleForTask checks score, skills and education gates.
func EligibleForTask(v domain.Volunteer, t domain.Task) error {
	if v.Status != domain.StatusApproved {
		return domain.ErrNotApproved
	}
	if t.MinScore > 0 && v.AverageScore+0.0001 < t.MinScore && v.CompletedTasks > 0 {
		return domain.ErrNotEligible
	}
	// New volunteers (no completed tasks) may take tasks with min score if they are approved.
	if t.MinScore > 0 && v.CompletedTasks > 0 && v.AverageScore < t.MinScore {
		return domain.ErrNotEligible
	}
	if !v.HasAnySkill(t.RequiredSkills) {
		return domain.ErrNotEligible
	}
	if t.RequiredEducation != "" && v.EducationField != t.RequiredEducation {
		return domain.ErrNotEligible
	}
	return nil
}

// RankingScore is a simple composite used for leaderboard ordering:
// hours first, then quality.
func RankingScore(hours, average float64, completed int) float64 {
	if completed == 0 {
		return hours
	}
	return hours*10 + average
}
