package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByExternalID(ctx context.Context, externalID string) (*User, error)
}

type VolunteerRepository interface {
	Create(ctx context.Context, v *Volunteer) error
	Update(ctx context.Context, v *Volunteer) error
	GetByID(ctx context.Context, id uuid.UUID) (*Volunteer, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Volunteer, error)
	List(ctx context.Context, f VolunteerFilter) ([]Volunteer, int, error)
	ReplaceAvailability(ctx context.Context, volunteerID uuid.UUID, slots []AvailabilitySlot) error
	ListAvailability(ctx context.Context, volunteerID uuid.UUID) ([]AvailabilitySlot, error)
	AddDocument(ctx context.Context, d *Document) error
	ListDocuments(ctx context.Context, volunteerID uuid.UUID) ([]Document, error)
	GetDocument(ctx context.Context, id uuid.UUID) (*Document, error)
	ReplaceSkills(ctx context.Context, volunteerID uuid.UUID, skillIDs []uuid.UUID) error
	ListVolunteerSkills(ctx context.Context, volunteerID uuid.UUID) ([]VolunteerSkill, error)
}

type SkillRepository interface {
	ListCatalog(ctx context.Context) ([]SkillGroup, error)
	CreateGroup(ctx context.Context, g *SkillGroup) error
	UpdateGroup(ctx context.Context, g *SkillGroup) error
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	CreateSkill(ctx context.Context, s *Skill) error
	UpdateSkill(ctx context.Context, s *Skill) error
	DeleteSkill(ctx context.Context, id uuid.UUID) error
	GetSkill(ctx context.Context, id uuid.UUID) (*Skill, error)
	GetSkillByTitle(ctx context.Context, groupID uuid.UUID, title string) (*Skill, error)
	GetGroup(ctx context.Context, id uuid.UUID) (*SkillGroup, error)
	CreateProposal(ctx context.Context, p *SkillProposal) error
	ListProposals(ctx context.Context, status SkillProposalStatus) ([]SkillProposal, error)
	ListProposalsByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]SkillProposal, error)
	GetProposal(ctx context.Context, id uuid.UUID) (*SkillProposal, error)
	UpdateProposal(ctx context.Context, p *SkillProposal) error
	SeedDefaults(ctx context.Context) error
}

type VolunteerFilter struct {
	Status VolunteerStatus
	Skill  SkillCategory
	Query  string
	Limit  int
	Offset int
}

type TaskRepository interface {
	Create(ctx context.Context, t *Task) error
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	List(ctx context.Context, f TaskFilter) ([]Task, int, error)
	ApplySeat(ctx context.Context, taskID, volunteerID uuid.UUID) (*Assignment, error)
	ReserveSeat(ctx context.Context, taskID, volunteerID uuid.UUID) (*Assignment, error)
	GetAssignment(ctx context.Context, id uuid.UUID) (*Assignment, error)
	GetAssignmentByTaskVolunteer(ctx context.Context, taskID, volunteerID uuid.UUID) (*Assignment, error)
	UpdateAssignment(ctx context.Context, a *Assignment) error
	ListAssignments(ctx context.Context, f AssignmentFilter) ([]Assignment, int, error)
}

type TaskFilter struct {
	Status             TaskStatus
	Skill              SkillCategory
	Query              string
	Upcoming           bool
	ExcludeVolunteerID uuid.UUID
	Limit              int
	Offset             int
}

type AssignmentFilter struct {
	VolunteerID uuid.UUID
	TaskID      uuid.UUID
	Status      AssignmentStatus
	Limit       int
	Offset      int
}

type MissionRepository interface {
	Create(ctx context.Context, m *Mission) error
	Update(ctx context.Context, m *Mission) error
	GetByID(ctx context.Context, id uuid.UUID) (*Mission, error)
	List(ctx context.Context, activeOnly bool) ([]Mission, error)
	GetByWebhookEvent(ctx context.Context, event string) ([]Mission, error)
	UpsertProgress(ctx context.Context, p *MissionProgress) error
	GetProgress(ctx context.Context, missionID, volunteerID uuid.UUID) (*MissionProgress, error)
	ListProgressByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]MissionProgress, error)
}

type CertificateRepository interface {
	Create(ctx context.Context, c *Certificate) error
	GetByVerificationCode(ctx context.Context, code uuid.UUID) (*Certificate, error)
	ListByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]Certificate, error)
	ExistsForAssignment(ctx context.Context, assignmentID uuid.UUID) (bool, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Notification, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
}

type StatsRepository interface {
	Dashboard(ctx context.Context) (*DashboardStats, error)
	Ranking(ctx context.Context, limit int) ([]RankingRow, error)
	SkillDistribution(ctx context.Context) (map[string]int, error)
}

type ObjectStorage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, string, error)
}

type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (func(), error)
}

type Notifier interface {
	Notify(ctx context.Context, userID uuid.UUID, title, body string) error
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }
