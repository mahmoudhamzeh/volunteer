package domain

import (
	"time"

	"github.com/google/uuid"
)

type MissionKind string

const (
	MissionCompleteProfile MissionKind = "complete_profile"
	MissionInviteUsers     MissionKind = "invite_users"
	MissionCustom          MissionKind = "custom"
	MissionWebhook         MissionKind = "webhook"
)

type MissionVerifyMode string

const (
	VerifyInternal MissionVerifyMode = "internal"
	VerifyOutbound MissionVerifyMode = "outbound"
	VerifyInbound  MissionVerifyMode = "inbound"
)

type MissionStatus string

const (
	MissionActive   MissionStatus = "active"
	MissionArchived MissionStatus = "archived"
)

type Mission struct {
	ID            uuid.UUID         `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Kind          MissionKind       `json:"kind"`
	HourWeight    float64           `json:"hour_weight"`
	DeadlineHours *int              `json:"deadline_hours,omitempty"`
	WebhookEvent  string            `json:"webhook_event,omitempty"`
	TargetCount   int               `json:"target_count"`
	VerifyMode    MissionVerifyMode `json:"verify_mode"`
	VerifyURL     string            `json:"verify_url,omitempty"`
	VerifyToken   string            `json:"verify_token,omitempty"`
	CanCheck      bool              `json:"can_check,omitempty"`
	Status        MissionStatus     `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
}

func (m Mission) Public() Mission {
	m.CanCheck = m.VerifyMode == VerifyInternal || m.VerifyMode == VerifyOutbound || (m.VerifyMode == VerifyInbound && m.VerifyURL != "")
	m.VerifyToken = ""
	m.VerifyURL = ""
	return m
}

type MissionProgressStatus string

const (
	MissionInProgress MissionProgressStatus = "in_progress"
	MissionCompleted  MissionProgressStatus = "completed"
	MissionExpired    MissionProgressStatus = "expired"
)

type MissionProgress struct {
	ID          uuid.UUID             `json:"id"`
	MissionID   uuid.UUID             `json:"mission_id"`
	VolunteerID uuid.UUID             `json:"volunteer_id"`
	Status      MissionProgressStatus `json:"status"`
	Progress    int                   `json:"progress"`
	StartedAt   time.Time             `json:"started_at"`
	DueAt       *time.Time            `json:"due_at,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	Mission     *Mission              `json:"mission,omitempty"`
}

type CertificateKind string

const (
	CertTask       CertificateKind = "task"
	CertAggregated CertificateKind = "aggregated"
)

type Certificate struct {
	ID               uuid.UUID       `json:"id"`
	VerificationCode uuid.UUID       `json:"verification_code"`
	VolunteerID      uuid.UUID       `json:"volunteer_id"`
	Kind             CertificateKind `json:"kind"`
	AssignmentID     *uuid.UUID      `json:"assignment_id,omitempty"`
	Title            string          `json:"title"`
	Hours            float64         `json:"hours"`
	PeriodStart      *time.Time      `json:"period_start,omitempty"`
	PeriodEnd        *time.Time      `json:"period_end,omitempty"`
	IssuedAt         time.Time       `json:"issued_at"`
	Volunteer        *Volunteer      `json:"volunteer,omitempty"`
}

type CertificateRequestStatus string

const (
	CertReqPending  CertificateRequestStatus = "pending"
	CertReqApproved CertificateRequestStatus = "approved"
	CertReqRejected CertificateRequestStatus = "rejected"
)

type CertificateRequest struct {
	ID              uuid.UUID                `json:"id"`
	VolunteerID     uuid.UUID                `json:"volunteer_id"`
	VolunteerName   string                   `json:"volunteer_name,omitempty"`
	Kind            CertificateKind          `json:"kind"`
	AssignmentID    *uuid.UUID               `json:"assignment_id,omitempty"`
	AssignmentTitle string                   `json:"assignment_title,omitempty"`
	Status          CertificateRequestStatus `json:"status"`
	AdminNote       string                   `json:"admin_note,omitempty"`
	CertificateID   *uuid.UUID               `json:"certificate_id,omitempty"`
	CreatedAt       time.Time                `json:"created_at"`
	ReviewedAt      *time.Time               `json:"reviewed_at,omitempty"`
}

type Notification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type DashboardStats struct {
	TotalVolunteers    int            `json:"total_volunteers"`
	PendingVolunteers  int            `json:"pending_volunteers"`
	ApprovedVolunteers int            `json:"approved_volunteers"`
	OnlineEstimate     int            `json:"online_estimate"`
	OpenTasks          int            `json:"open_tasks"`
	ActiveAssignments  int            `json:"active_assignments"`
	CompletedThisMonth int            `json:"completed_this_month"`
	ParticipationRate  float64        `json:"participation_rate"`
	TotalHours         float64        `json:"total_hours"`
	SkillDistribution  map[string]int `json:"skill_distribution"`
}

type RankingRow struct {
	VolunteerID    uuid.UUID       `json:"volunteer_id"`
	FullName       string          `json:"full_name"`
	City           string          `json:"city"`
	Skills         []SkillCategory `json:"skills"`
	AverageScore   float64         `json:"average_score"`
	TotalHours     float64         `json:"total_hours"`
	CompletedTasks int             `json:"completed_tasks"`
	Status         VolunteerStatus `json:"status"`
}
