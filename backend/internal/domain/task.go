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
	TaskInactive  TaskStatus = "inactive"
)

const (
	WorkOnsite = "onsite"
	WorkRemote = "remote"
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
	WorkMode          string          `json:"work_mode"`
	DeliveryHint      string          `json:"delivery_hint"`
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

func (t Task) IsRemote() bool {
	return t.WorkMode == WorkRemote
}

func ParseWorkMode(s string) string {
	if s == WorkRemote {
		return WorkRemote
	}
	return WorkOnsite
}

func ValidTaskStatus(s TaskStatus) bool {
	switch s {
	case TaskOpen, TaskClosed, TaskCancelled, TaskInactive:
		return true
	default:
		return false
	}
}

type AssignmentStatus string

const (
	AssignmentRequested  AssignmentStatus = "requested"
	AssignmentReserved   AssignmentStatus = "reserved"
	AssignmentInProgress AssignmentStatus = "in_progress"
	AssignmentAttended   AssignmentStatus = "attended"
	AssignmentSubmitted  AssignmentStatus = "submitted"
	AssignmentCompleted  AssignmentStatus = "completed"
	AssignmentCancelled  AssignmentStatus = "cancelled"
	AssignmentRejected   AssignmentStatus = "rejected"
)

func (s AssignmentStatus) BlocksReapply() bool {
	switch s {
	case AssignmentRequested, AssignmentReserved, AssignmentInProgress, AssignmentAttended, AssignmentSubmitted, AssignmentCompleted:
		return true
	default:
		return false
	}
}

func (s AssignmentStatus) OccupiesSeat() bool {
	switch s {
	case AssignmentReserved, AssignmentInProgress, AssignmentAttended, AssignmentSubmitted, AssignmentCompleted:
		return true
	default:
		return false
	}
}

func (s AssignmentStatus) Cancellable() bool {
	return s == AssignmentRequested || s == AssignmentReserved || s == AssignmentInProgress || s == AssignmentSubmitted
}

type Assignment struct {
	ID                uuid.UUID        `json:"id"`
	TaskID            uuid.UUID        `json:"task_id"`
	VolunteerID       uuid.UUID        `json:"volunteer_id"`
	Status            AssignmentStatus `json:"status"`
	VolunteerRating   *int             `json:"volunteer_rating,omitempty"`
	VolunteerComment  string           `json:"volunteer_comment,omitempty"`
	AdminDiscipline   *int             `json:"admin_discipline,omitempty"`
	AdminExpertise    *int             `json:"admin_expertise,omitempty"`
	AdminEthics       *int             `json:"admin_ethics,omitempty"`
	AdminComment      string           `json:"admin_comment,omitempty"`
	CompositeScore    *float64         `json:"composite_score,omitempty"`
	HoursAwarded      float64          `json:"hours_awarded"`
	AttendedAt        *time.Time       `json:"attended_at,omitempty"`
	CompletedAt       *time.Time       `json:"completed_at,omitempty"`
	DeliveryNote      string           `json:"delivery_note,omitempty"`
	DeliveryFileName  string           `json:"delivery_file_name,omitempty"`
	DeliveryObjectKey string           `json:"-"`
	DeliveryMime      string           `json:"-"`
	DeliveredAt       *time.Time       `json:"delivered_at,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`

	Task      *Task      `json:"task,omitempty"`
	Volunteer *Volunteer `json:"volunteer,omitempty"`
}

func (a Assignment) CanIssueCertificate() bool {
	return a.Status == AssignmentCompleted && a.HoursAwarded > 0
}
