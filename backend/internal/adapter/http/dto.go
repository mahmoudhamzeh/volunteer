package httpserver

import (
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/missionuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/taskuc"
)

func userDTO(u *domain.User) map[string]any {
	return map[string]any{
		"id":               u.ID,
		"email":            u.Email,
		"role":             u.Role,
		"external_user_id": u.ExternalUserID,
		"created_at":       u.CreatedAt,
	}
}

func volunteerDTO(v *domain.Volunteer) map[string]any {
	return map[string]any{
		"id":               v.ID,
		"user_id":          v.UserID,
		"full_name":        v.FullName,
		"national_id":      v.NationalID,
		"phone":            v.Phone,
		"city":             v.City,
		"bio":              v.Bio,
		"skill_categories": v.SkillCategories,
		"education_field":  v.EducationField,
		"medical_license":  v.MedicalLicense,
		"status":           v.Status,
		"rejection_reason": v.RejectionReason,
		"average_score":    v.AverageScore,
		"total_hours":      v.TotalHours,
		"completed_tasks":  v.CompletedTasks,
		"created_at":       v.CreatedAt,
		"updated_at":       v.UpdatedAt,
	}
}

func certDTO(c *domain.Certificate) map[string]any {
	name := ""
	if c.Volunteer != nil {
		name = c.Volunteer.FullName
	}
	return map[string]any{
		"id":                c.ID,
		"verification_code": c.VerificationCode,
		"volunteer_id":      c.VolunteerID,
		"volunteer_name":    name,
		"kind":              c.Kind,
		"title":             c.Title,
		"hours":             c.Hours,
		"period_start":      c.PeriodStart,
		"period_end":        c.PeriodEnd,
		"issued_at":         c.IssuedAt,
		"authentic":         true,
	}
}

func taskInput(in taskBody) taskuc.TaskInput {
	return taskuc.TaskInput{
		Title:             in.Title,
		Description:       in.Description,
		Location:          in.Location,
		StartsAt:          in.StartsAt,
		EndsAt:            in.EndsAt,
		Capacity:          in.Capacity,
		HourWeight:        in.HourWeight,
		RequiredSkills:    in.RequiredSkills,
		MinScore:          in.MinScore,
		RequiredEducation: in.RequiredEducation,
		Status:            domain.TaskStatus(in.Status),
	}
}

func missionIn(title, desc, kind string, hours float64, deadline *int, event string, target int) missionuc.MissionInput {
	return missionuc.MissionInput{
		Title:         title,
		Description:   desc,
		Kind:          domain.MissionKind(kind),
		HourWeight:    hours,
		DeadlineHours: deadline,
		WebhookEvent:  event,
		TargetCount:   target,
	}
}
