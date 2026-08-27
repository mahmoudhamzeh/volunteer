package domain

import (
	"strings"
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

const (
	TrainingInPerson = "in_person"
	TrainingOnline   = "online"
	TrainingHybrid   = "hybrid"
	TrainingWorkshop = "workshop"
	TrainingOther    = "other"
)

const (
	TaskOneOff     = "one_off"
	TaskRecurring  = "recurring"
	TaskOccurrence = "occurrence"
)

type TaskSlot struct {
	Weekday   int    `json:"weekday"`
	Capacity  int    `json:"capacity"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

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
	RequiresTraining  bool            `json:"requires_training"`
	TrainingCourseID  uuid.UUID       `json:"training_course_id,omitempty"`
	TrainingCourse    *TrainingCourse `json:"training_course,omitempty"`
	TrainingKind      string          `json:"training_kind,omitempty"`
	TrainingLocation  string          `json:"training_location,omitempty"`
	TrainingAt        *time.Time      `json:"training_at,omitempty"`
	Kind              string          `json:"kind"`
	SeriesID          uuid.UUID       `json:"series_id,omitempty"`
	Weekday           int             `json:"weekday"`
	Slots             []TaskSlot      `json:"slots,omitempty"`
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

func ValidTrainingKind(s string) bool {
	switch s {
	case TrainingInPerson, TrainingOnline, TrainingHybrid, TrainingWorkshop, TrainingOther:
		return true
	default:
		return false
	}
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
	AssignmentRequested         AssignmentStatus = "requested"
	AssignmentTrainingPending   AssignmentStatus = "training_pending"
	AssignmentReserved          AssignmentStatus = "reserved"
	AssignmentInProgress        AssignmentStatus = "in_progress"
	AssignmentAttended          AssignmentStatus = "attended"
	AssignmentSubmitted         AssignmentStatus = "submitted"
	AssignmentCompleted         AssignmentStatus = "completed"
	AssignmentCancelled         AssignmentStatus = "cancelled"
	AssignmentRejected          AssignmentStatus = "rejected"
	AssignmentAbsent            AssignmentStatus = "absent"
	AssignmentRevisionRequested AssignmentStatus = "revision_requested"
)

func (s AssignmentStatus) BlocksReapply() bool {
	switch s {
	case AssignmentRequested, AssignmentTrainingPending, AssignmentReserved, AssignmentInProgress, AssignmentAttended, AssignmentSubmitted, AssignmentCompleted, AssignmentRevisionRequested:
		return true
	default:
		return false
	}
}

func (s AssignmentStatus) OccupiesSeat() bool {
	switch s {
	case AssignmentTrainingPending, AssignmentReserved, AssignmentInProgress, AssignmentAttended, AssignmentSubmitted, AssignmentCompleted, AssignmentRevisionRequested:
		return true
	default:
		return false
	}
}

func (s AssignmentStatus) Cancellable() bool {
	return s == AssignmentRequested || s == AssignmentTrainingPending || s == AssignmentReserved || s == AssignmentInProgress || s == AssignmentSubmitted || s == AssignmentRevisionRequested
}

// TrainingSeriesID is the recurring family this activity belongs to, if any.
func (t Task) TrainingSeriesID() uuid.UUID {
	if t.SeriesID != uuid.Nil {
		return t.SeriesID
	}
	if t.Kind == TaskRecurring {
		return t.ID
	}
	return uuid.Nil
}

const (
	TrainingCourseActive   = "active"
	TrainingCourseInactive = "inactive"
)

type TrainingCourse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Kind        string     `json:"kind"`
	Location    string     `json:"location"`
	TrainingAt  *time.Time `json:"training_at,omitempty"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (c TrainingCourse) IsActive() bool {
	return c.Status == "" || c.Status == TrainingCourseActive
}

func (c TrainingCourse) DisplayTitle() string {
	if t := strings.TrimSpace(c.Title); t != "" {
		return t
	}
	return "دوره آموزشی"
}

// CoversTask reports whether this completed course satisfies another activity's training requirement.
func (vt VolunteerTraining) CoversTask(t Task) bool {
	if t.TrainingCourseID != uuid.Nil && vt.CourseID == t.TrainingCourseID {
		return true
	}
	title := strings.ToLower(strings.TrimSpace(t.TrainingCourseTitle()))
	if title != "" && strings.ToLower(strings.TrimSpace(vt.CourseTitle)) == title {
		return true
	}
	sid := t.TrainingSeriesID()
	if sid != uuid.Nil && vt.SeriesID == sid {
		return true
	}
	kind := strings.ToLower(strings.TrimSpace(t.TrainingKind))
	loc := strings.ToLower(strings.TrimSpace(t.TrainingLocation))
	if kind == "" || loc == "" {
		return false
	}
	return strings.ToLower(strings.TrimSpace(vt.TrainingKind)) == kind &&
		strings.ToLower(strings.TrimSpace(vt.TrainingLocation)) == loc
}

func (t Task) TrainingCourseTitle() string {
	if t.TrainingCourse != nil && strings.TrimSpace(t.TrainingCourse.Title) != "" {
		return t.TrainingCourse.Title
	}
	return ""
}

func (t *Task) ApplyCourse(c *TrainingCourse) {
	if t == nil {
		return
	}
	if c == nil || c.ID == uuid.Nil {
		t.TrainingCourseID = uuid.Nil
		t.TrainingCourse = nil
		t.TrainingKind = ""
		t.TrainingLocation = ""
		t.TrainingAt = nil
		return
	}
	cp := *c
	t.TrainingCourseID = c.ID
	t.TrainingCourse = &cp
	t.TrainingKind = c.Kind
	t.TrainingLocation = c.Location
	t.TrainingAt = c.TrainingAt
}

type VolunteerTraining struct {
	ID               uuid.UUID  `json:"id"`
	VolunteerID      uuid.UUID  `json:"volunteer_id"`
	CourseID         uuid.UUID  `json:"course_id,omitempty"`
	CourseTitle      string     `json:"course_title,omitempty"`
	SeriesID         uuid.UUID  `json:"series_id,omitempty"`
	TrainingKind     string     `json:"training_kind"`
	TrainingLocation string     `json:"training_location"`
	TrainingAt       *time.Time `json:"training_at,omitempty"`
	SourceTaskID     uuid.UUID  `json:"source_task_id,omitempty"`
	SourceTaskTitle  string     `json:"source_task_title,omitempty"`
	AssignmentID     uuid.UUID  `json:"assignment_id,omitempty"`
	ConfirmedBy      uuid.UUID  `json:"confirmed_by,omitempty"`
	ConfirmedAt      time.Time  `json:"confirmed_at"`
}

func (vt VolunteerTraining) DisplayTitle() string {
	if strings.TrimSpace(vt.CourseTitle) != "" {
		return vt.CourseTitle
	}
	if strings.TrimSpace(vt.SourceTaskTitle) != "" {
		return vt.SourceTaskTitle
	}
	return "دوره آموزشی"
}

type Assignment struct {
	ID                uuid.UUID         `json:"id"`
	TaskID            uuid.UUID         `json:"task_id"`
	VolunteerID       uuid.UUID         `json:"volunteer_id"`
	Status            AssignmentStatus  `json:"status"`
	VolunteerRating   *int              `json:"volunteer_rating,omitempty"`
	VolunteerComment  string            `json:"volunteer_comment,omitempty"`
	AdminDiscipline   *int              `json:"admin_discipline,omitempty"`
	AdminExpertise    *int              `json:"admin_expertise,omitempty"`
	AdminEthics       *int              `json:"admin_ethics,omitempty"`
	AdminComment      string            `json:"admin_comment,omitempty"`
	CompositeScore    *float64          `json:"composite_score,omitempty"`
	HoursAwarded      float64           `json:"hours_awarded"`
	AttendedAt        *time.Time        `json:"attended_at,omitempty"`
	CheckInAt         *time.Time        `json:"check_in_at,omitempty"`
	CheckOutAt        *time.Time        `json:"check_out_at,omitempty"`
	CompletedAt       *time.Time        `json:"completed_at,omitempty"`
	DeliveryNote      string            `json:"delivery_note,omitempty"`
	DeliveryFileName  string            `json:"delivery_file_name,omitempty"`
	DeliveryObjectKey string            `json:"-"`
	DeliveryMime      string            `json:"-"`
	DeliveredAt       *time.Time        `json:"delivered_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	History           []AssignmentEvent `json:"history,omitempty"`

	Task      *Task      `json:"task,omitempty"`
	Volunteer *Volunteer `json:"volunteer,omitempty"`
}

const (
	AssignmentEventDelivery = "delivery"
	AssignmentEventRevision = "revision"
	AssignmentEventMessage  = "message"
)

type AssignmentEventFile struct {
	ID           uuid.UUID `json:"id"`
	EventID      uuid.UUID `json:"event_id"`
	AssignmentID uuid.UUID `json:"assignment_id,omitempty"`
	FileName     string    `json:"file_name"`
	ObjectKey    string    `json:"-"`
	MimeType     string    `json:"mime_type,omitempty"`
	SizeBytes    int64     `json:"size_bytes,omitempty"`
}

type AssignmentEvent struct {
	ID           uuid.UUID             `json:"id"`
	AssignmentID uuid.UUID             `json:"assignment_id"`
	Kind         string                `json:"kind"`
	Note         string                `json:"note"`
	ActorRole    string                `json:"actor_role"`
	CreatedAt    time.Time             `json:"created_at"`
	Files        []AssignmentEventFile `json:"files,omitempty"`
}

func (a Assignment) CanIssueCertificate() bool {
	return a.Status == AssignmentCompleted
}
