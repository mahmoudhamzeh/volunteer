package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type VolunteerStatus string

const (
	StatusDraft     VolunteerStatus = "draft"
	StatusPending   VolunteerStatus = "pending"
	StatusRejected  VolunteerStatus = "rejected"
	StatusApproved  VolunteerStatus = "approved"
	StatusSuspended VolunteerStatus = "suspended"
)

func (s VolunteerStatus) CanViewTasks() bool {
	return s == StatusApproved
}

type SkillCategory string

const (
	SkillMedical        SkillCategory = "medical"
	SkillAdministrative SkillCategory = "administrative"
	SkillArtistic       SkillCategory = "artistic"
	SkillTechnical      SkillCategory = "technical"
	SkillEducation      SkillCategory = "education"
	SkillLogistics      SkillCategory = "logistics"
	SkillPsychological  SkillCategory = "psychological"
)

var AllSkillCategories = []SkillCategory{
	SkillMedical, SkillAdministrative, SkillArtistic, SkillTechnical,
	SkillEducation, SkillLogistics, SkillPsychological,
}

type Volunteer struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	FullName        string          `json:"full_name"`
	NationalID      string          `json:"national_id"`
	Phone           string          `json:"phone"`
	City            string          `json:"city"`
	Bio             string          `json:"bio"`
	SkillCategories []SkillCategory `json:"skill_categories"`
	EducationField  string          `json:"education_field"`
	MedicalLicense  string          `json:"medical_license"`
	Status          VolunteerStatus `json:"status"`
	RejectionReason string          `json:"rejection_reason"`
	AverageScore    float64         `json:"average_score"`
	TotalHours      float64         `json:"total_hours"`
	CompletedTasks  int             `json:"completed_tasks"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (v Volunteer) HasSkill(skill SkillCategory) bool {
	for _, s := range v.SkillCategories {
		if s == skill {
			return true
		}
	}
	return false
}

func (v Volunteer) HasAnySkill(required []SkillCategory) bool {
	if len(required) == 0 {
		return true
	}
	set := map[SkillCategory]struct{}{}
	for _, s := range v.SkillCategories {
		set[s] = struct{}{}
	}
	for _, r := range required {
		if _, ok := set[r]; ok {
			return true
		}
	}
	return false
}

type AvailabilitySlot struct {
	ID          uuid.UUID `json:"id"`
	VolunteerID uuid.UUID `json:"volunteer_id"`
	Weekday     int       `json:"weekday"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
}

type DocumentKind string

const (
	DocNationalID     DocumentKind = "national_id"
	DocDrivingLicense DocumentKind = "driving_license"
	DocMedicalLicense DocumentKind = "medical_license"
	DocEducation      DocumentKind = "education"
	DocOther          DocumentKind = "other"
)

type Document struct {
	ID          uuid.UUID    `json:"id"`
	VolunteerID uuid.UUID    `json:"volunteer_id"`
	Kind        DocumentKind `json:"kind"`
	ObjectKey   string       `json:"-"`
	FileName    string       `json:"file_name"`
	MimeType    string       `json:"mime_type"`
	SizeBytes   int64        `json:"size_bytes"`
	CreatedAt   time.Time    `json:"created_at"`
}

func ParseSkillCategories(values []string) []SkillCategory {
	out := make([]SkillCategory, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" {
			continue
		}
		out = append(out, SkillCategory(v))
	}
	return out
}

// CanTransition validates the volunteer identity state machine.
func CanTransition(from, to VolunteerStatus) bool {
	switch from {
	case StatusDraft:
		return to == StatusPending
	case StatusPending:
		return to == StatusApproved || to == StatusRejected || to == StatusDraft
	case StatusRejected:
		return to == StatusPending || to == StatusDraft
	case StatusApproved:
		return to == StatusSuspended
	case StatusSuspended:
		return to == StatusApproved
	default:
		return false
	}
}
