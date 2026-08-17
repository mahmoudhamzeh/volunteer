package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskOpen      TaskStatus = "open"
	TaskClosed    TaskStatus = "closed"
	TaskCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID                uuid.UUID       `json:"id"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	Location          string          `json:"location"`
	StartsAt          time.Time       `json:"starts_at"`
	EndsAt            time.Time       `json:"ends_at"`
	Capacity          int             `json:"capacity"`
	ReservedCount     int             `json:"reserved_count"`
	HourWeight        float64         `json:"hour_weight"`
	RequiredSkills    []SkillCategory `json:"required_skills"`
	RequiredSkillIDs  []uuid.UUID     `json:"required_skill_ids"`
	MinScore          float64         `json:"min_score"`
	RequiredEducation string          `json:"required_education"`
	Status            TaskStatus      `json:"status"`
	CreatedBy         uuid.UUID       `json:"created_by"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func (t Task) RemainingCapacity() int {
	left := t.Capacity - t.ReservedCount
	if left < 0 {
		return 0
	}
	return left
}

func (t Task) IsOpen() bool {
	return t.Status == TaskOpen && time.Now().Before(t.EndsAt)
}

type AssignmentStatus string

const (
	AssignmentReserved  AssignmentStatus = "reserved"
	AssignmentAttended  AssignmentStatus = "attended"
	AssignmentCompleted AssignmentStatus = "completed"
	AssignmentCancelled AssignmentStatus = "cancelled"
	AssignmentRejected  AssignmentStatus = "rejected"
)

type Assignment struct {
	ID               uuid.UUID        `json:"id"`
	TaskID           uuid.UUID        `json:"task_id"`
	VolunteerID      uuid.UUID        `json:"volunteer_id"`
	Status           AssignmentStatus `json:"status"`
	VolunteerRating  *int             `json:"volunteer_rating,omitempty"`
	VolunteerComment string           `json:"volunteer_comment,omitempty"`
	AdminDiscipline  *int             `json:"admin_discipline,omitempty"`
	AdminExpertise   *int             `json:"admin_expertise,omitempty"`
	AdminEthics      *int             `json:"admin_ethics,omitempty"`
	AdminComment     string           `json:"admin_comment,omitempty"`
	CompositeScore   *float64         `json:"composite_score,omitempty"`
	HoursAwarded     float64          `json:"hours_awarded"`
	AttendedAt       *time.Time       `json:"attended_at,omitempty"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`

	Task      *Task      `json:"task,omitempty"`
	Volunteer *Volunteer `json:"volunteer,omitempty"`
}

func (a Assignment) CanIssueCertificate() bool {
	return a.Status == AssignmentCompleted && a.HoursAwarded > 0
}
