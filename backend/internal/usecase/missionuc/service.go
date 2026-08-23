package missionuc

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type Service struct {
	missions   domain.MissionRepository
	volunteers domain.VolunteerRepository
	notify     domain.Notifier
	clock      domain.Clock
}

func New(missions domain.MissionRepository, volunteers domain.VolunteerRepository, notify domain.Notifier, clock domain.Clock) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	return &Service{missions: missions, volunteers: volunteers, notify: notify, clock: clock}
}

type MissionInput struct {
	Title         string
	Description   string
	Kind          domain.MissionKind
	HourWeight    float64
	DeadlineHours *int
	WebhookEvent  string
	TargetCount   int
	Status        domain.MissionStatus
}

func (s *Service) Create(ctx context.Context, in MissionInput) (*domain.Mission, error) {
	if strings.TrimSpace(in.Title) == "" || in.HourWeight <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if in.Kind == "" {
		in.Kind = domain.MissionCustom
	}
	m := &domain.Mission{
		ID:            uuid.New(),
		Title:         strings.TrimSpace(in.Title),
		Description:   strings.TrimSpace(in.Description),
		Kind:          in.Kind,
		HourWeight:    in.HourWeight,
		DeadlineHours: in.DeadlineHours,
		WebhookEvent:  strings.TrimSpace(in.WebhookEvent),
		TargetCount:   in.TargetCount,
		Status:        domain.MissionActive,
		CreatedAt:     s.clock.Now(),
	}
	if in.Status != "" {
		m.Status = in.Status
	}
	if m.TargetCount <= 0 {
		m.TargetCount = 1
	}
	return m, s.missions.Create(ctx, m)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in MissionInput) (*domain.Mission, error) {
	m, err := s.missions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Title) != "" {
		m.Title = strings.TrimSpace(in.Title)
	}
	m.Description = strings.TrimSpace(in.Description)
	if in.Kind != "" {
		m.Kind = in.Kind
	}
	if in.HourWeight > 0 {
		m.HourWeight = in.HourWeight
	}
	m.DeadlineHours = in.DeadlineHours
	m.WebhookEvent = strings.TrimSpace(in.WebhookEvent)
	if in.TargetCount > 0 {
		m.TargetCount = in.TargetCount
	}
	if in.Status != "" {
		m.Status = in.Status
	}
	return m, s.missions.Update(ctx, m)
}

func (s *Service) List(ctx context.Context, activeOnly bool) ([]domain.Mission, error) {
	return s.missions.List(ctx, activeOnly)
}

func (s *Service) Start(ctx context.Context, userID, missionID uuid.UUID) (*domain.MissionProgress, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !v.Status.CanViewTasks() && v.Status != domain.StatusDraft && v.Status != domain.StatusPending {
		return nil, domain.ErrNotApproved
	}
	m, err := s.missions.GetByID(ctx, missionID)
	if err != nil {
		return nil, err
	}
	if m.Status != domain.MissionActive {
		return nil, domain.ErrNotEligible
	}
	now := s.clock.Now()
	p := &domain.MissionProgress{
		ID:          uuid.New(),
		MissionID:   m.ID,
		VolunteerID: v.ID,
		Status:      domain.MissionInProgress,
		Progress:    0,
		StartedAt:   now,
		Mission:     m,
	}
	if m.DeadlineHours != nil {
		due := now.Add(time.Duration(*m.DeadlineHours) * time.Hour)
		p.DueAt = &due
	}
	if err := s.missions.UpsertProgress(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) MyProgress(ctx context.Context, userID uuid.UUID) ([]domain.MissionProgress, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.missions.ListProgressByVolunteer(ctx, v.ID)
}

func (s *Service) ReportProgress(ctx context.Context, userID, missionID uuid.UUID, increment int) (*domain.MissionProgress, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.advance(ctx, v, missionID, increment)
}

func (s *Service) AwardWebhook(ctx context.Context, volunteerID uuid.UUID, event string, increment int) error {
	if increment <= 0 {
		increment = 1
	}
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return err
	}
	missions, err := s.missions.GetByWebhookEvent(ctx, event)
	if err != nil {
		return err
	}
	for _, m := range missions {
		if _, err := s.advance(ctx, v, m.ID, increment); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) advance(ctx context.Context, v *domain.Volunteer, missionID uuid.UUID, increment int) (*domain.MissionProgress, error) {
	m, err := s.missions.GetByID(ctx, missionID)
	if err != nil {
		return nil, err
	}
	p, err := s.missions.GetProgress(ctx, missionID, v.ID)
	now := s.clock.Now()
	if err == domain.ErrNotFound {
		p = &domain.MissionProgress{
			ID:          uuid.New(),
			MissionID:   missionID,
			VolunteerID: v.ID,
			Status:      domain.MissionInProgress,
			StartedAt:   now,
		}
		if m.DeadlineHours != nil {
			due := now.Add(time.Duration(*m.DeadlineHours) * time.Hour)
			p.DueAt = &due
		}
	} else if err != nil {
		return nil, err
	}
	if p.Status == domain.MissionCompleted {
		p.Mission = m
		return p, nil
	}
	if p.DueAt != nil && now.After(*p.DueAt) {
		p.Status = domain.MissionExpired
		_ = s.missions.UpsertProgress(ctx, p)
		return nil, domain.ErrMissionExpired
	}
	p.Progress += increment
	if p.Progress >= m.TargetCount {
		p.Progress = m.TargetCount
		p.Status = domain.MissionCompleted
		p.CompletedAt = &now
		updated, err := s.volunteers.AddCompletedWork(ctx, v.ID, 5, m.HourWeight)
		if err != nil {
			return nil, err
		}
		v = updated
		if s.notify != nil {
			_ = s.notify.Notify(ctx, v.UserID, "ماموریت تکمیل شد", "امتیاز و ساعات داوطلبی ماموریت «"+m.Title+"» به حساب شما افزوده شد.")
		}
	}
	if err := s.missions.UpsertProgress(ctx, p); err != nil {
		return nil, err
	}
	p.Mission = m
	return p, nil
}
