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
	SkillSports         SkillCategory = "sports"
	SkillPsychological  SkillCategory = "psychological"
)

var AllSkillCategories = []SkillCategory{
	SkillMedical, SkillAdministrative, SkillArtistic, SkillTechnical,
	SkillEducation, SkillLogistics, SkillPsychological, SkillSports,
}

type Volunteer struct {
	ID              uuid.UUID        `json:"id"`
	UserID          uuid.UUID        `json:"user_id"`
	FullName        string           `json:"full_name"`
	FirstName       string           `json:"first_name"`
	LastName        string           `json:"last_name"`
	NationalID      string           `json:"national_id"`
	Phone           string           `json:"phone"`
	Phone2          string           `json:"phone2"`
	Province        string           `json:"province"`
	City            string           `json:"city"`
	Address         string           `json:"address"`
	Plaque          string           `json:"plaque"`
	Unit            string           `json:"unit"`
	Bio             string           `json:"bio"`
	SkillCategories []SkillCategory  `json:"skill_categories"`
	Skills          []VolunteerSkill `json:"skills"`
	Proposals       []SkillProposal  `json:"proposals"`
	EducationLevel  string           `json:"education_level"`
	EducationField  string           `json:"education_field"`
	MedicalLicense  string           `json:"medical_license"`
	BirthDate       string           `json:"birth_date"`
	Status          VolunteerStatus  `json:"status"`
	RejectionReason string           `json:"rejection_reason"`
	Email           string           `json:"email"`
	AverageScore    float64          `json:"average_score"`
	TotalHours      float64          `json:"total_hours"`
	CompletedTasks  int              `json:"completed_tasks"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	History         []VolunteerEvent `json:"history,omitempty"`
}

type VolunteerEventType string

const (
	EventSubmitted          VolunteerEventType = "submitted"
	EventApproved           VolunteerEventType = "approved"
	EventRejected           VolunteerEventType = "rejected"
	EventDocumentsRequested VolunteerEventType = "documents_requested"
	EventSuspended          VolunteerEventType = "suspended"
	EventUnsuspended        VolunteerEventType = "unsuspended"
	EventStatusChanged      VolunteerEventType = "status_changed"
	EventComment            VolunteerEventType = "comment"
	EventProfileUpdated     VolunteerEventType = "profile_updated"
	EventDocumentDeleted    VolunteerEventType = "document_deleted"
)

type VolunteerEvent struct {
	ID          uuid.UUID          `json:"id"`
	VolunteerID uuid.UUID          `json:"volunteer_id"`
	ActorUserID uuid.UUID          `json:"actor_user_id"`
	ActorRole   string             `json:"actor_role"`
	EventType   VolunteerEventType `json:"event_type"`
	FromStatus  VolunteerStatus    `json:"from_status"`
	ToStatus    VolunteerStatus    `json:"to_status"`
	Comment     string             `json:"comment"`
	CreatedAt   time.Time          `json:"created_at"`
}

func ParseVolunteerStatus(s string) (VolunteerStatus, bool) {
	switch VolunteerStatus(s) {
	case StatusDraft, StatusPending, StatusRejected, StatusApproved, StatusSuspended:
		return VolunteerStatus(s), true
	default:
		return "", false
	}
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

func (v Volunteer) HasAnySkillID(ids []uuid.UUID) bool {
	if len(ids) == 0 {
		return true
	}
	set := map[uuid.UUID]struct{}{}
	for _, s := range v.Skills {
		set[s.SkillID] = struct{}{}
	}
	for _, id := range ids {
		if _, ok := set[id]; ok {
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
