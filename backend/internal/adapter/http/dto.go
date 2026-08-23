package httpserver

import (
	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/missionuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/taskuc"
)

func userDTO(u *domain.User) map[string]any {
	return map[string]any{
		"id":               u.ID,
		"email":            u.Email,
		"phone":            u.Phone,
		"role":             u.Role,
		"external_user_id": u.ExternalUserID,
		"created_at":       u.CreatedAt,
	}
}

func volunteerDTO(v *domain.Volunteer) map[string]any {
	ids := make([]uuid.UUID, 0, len(v.Skills))
	for _, s := range v.Skills {
		ids = append(ids, s.SkillID)
	}
	return map[string]any{
		"id":               v.ID,
		"user_id":          v.UserID,
		"full_name":        v.FullName,
		"first_name":       v.FirstName,
		"last_name":        v.LastName,
		"national_id":      v.NationalID,
		"phone":            v.Phone,
		"phone2":           v.Phone2,
		"province":         v.Province,
		"city":             v.City,
		"address":          v.Address,
		"plaque":           v.Plaque,
		"unit":             v.Unit,
		"bio":              v.Bio,
		"skill_categories": nonempty(v.SkillCategories),
		"skill_ids":        nonempty(ids),
		"skills":           nonempty(v.Skills),
		"proposals":        nonempty(v.Proposals),
		"education_level":  v.EducationLevel,
		"education_field":  v.EducationField,
		"medical_license":  v.MedicalLicense,
		"birth_date":       v.BirthDate,
		"gender":           v.Gender,
		"occupation":       v.Occupation,
		"occupation_other": v.OccupationOther,
		"email":            v.Email,
		"status":           v.Status,
		"rejection_reason": v.RejectionReason,
		"history":          nonempty(v.History),
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
		RequiredSkillIDs:  in.RequiredSkillIDs,
		MinScore:          in.MinScore,
		RequiredEducation: in.RequiredEducation,
		WorkMode:          in.WorkMode,
		DeliveryHint:      in.DeliveryHint,
		Status:            domain.TaskStatus(in.Status),
		Kind:              in.Kind,
		Slots:             in.Slots,
	}
}

func nonempty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func missionIn(title, desc, kind string, hours float64, deadline *int, event string, target int, mode, url, token string) missionuc.MissionInput {
	return missionuc.MissionInput{
		Title:         title,
		Description:   desc,
		Kind:          domain.MissionKind(kind),
		HourWeight:    hours,
		DeadlineHours: deadline,
		WebhookEvent:  event,
		TargetCount:   target,
		VerifyMode:    domain.MissionVerifyMode(mode),
		VerifyURL:     url,
		VerifyToken:   token,
	}
}
