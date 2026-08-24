package missionuc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/scoring"
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Service struct {
	missions   domain.MissionRepository
	volunteers domain.VolunteerRepository
	notify     domain.Notifier
	clock      domain.Clock
	http       HTTPDoer
}

func New(missions domain.MissionRepository, volunteers domain.VolunteerRepository, notify domain.Notifier, clock domain.Clock, client HTTPDoer) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &Service{missions: missions, volunteers: volunteers, notify: notify, clock: clock, http: client}
}

type MissionInput struct {
	Title         string
	Description   string
	Kind          domain.MissionKind
	HourWeight    float64
	DeadlineHours *int
	WebhookEvent  string
	TargetCount   int
	VerifyMode    domain.MissionVerifyMode
	VerifyURL     string
	VerifyToken   string
	Status        domain.MissionStatus
}

func defaultVerifyMode(kind domain.MissionKind) domain.MissionVerifyMode {
	switch kind {
	case domain.MissionCompleteProfile:
		return domain.VerifyInternal
	case domain.MissionInviteUsers, domain.MissionWebhook:
		return domain.VerifyInbound
	default:
		return domain.VerifyOutbound
	}
}

func newVerifyToken() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return strings.ReplaceAll(uuid.NewString()+uuid.NewString(), "-", "")
	}
	return hex.EncodeToString(b)
}

func (s *Service) normalize(in *MissionInput, existing *domain.Mission) error {
	if in.Kind == "" && existing != nil {
		in.Kind = existing.Kind
	}
	if in.Kind == "" {
		in.Kind = domain.MissionCustom
	}
	if in.VerifyMode == "" {
		if existing != nil && existing.VerifyMode != "" {
			in.VerifyMode = existing.VerifyMode
		} else {
			in.VerifyMode = defaultVerifyMode(in.Kind)
		}
	}
	in.VerifyURL = strings.TrimSpace(in.VerifyURL)
	in.VerifyToken = strings.TrimSpace(in.VerifyToken)
	in.WebhookEvent = strings.TrimSpace(in.WebhookEvent)
	if in.Kind == domain.MissionCompleteProfile {
		in.VerifyMode = domain.VerifyInternal
	}
	if in.VerifyMode == domain.VerifyOutbound && in.VerifyURL == "" {
		if existing != nil {
			in.VerifyURL = strings.TrimSpace(existing.VerifyURL)
		}
		if in.VerifyURL == "" {
			return domain.Invalid("برای تأیید از طریق وب‌سرویس، آدرس سرویس را وارد کنید")
		}
	}
	if in.VerifyMode != domain.VerifyInternal && in.VerifyToken == "" {
		if existing != nil {
			in.VerifyToken = strings.TrimSpace(existing.VerifyToken)
		}
		if in.VerifyToken == "" {
			in.VerifyToken = newVerifyToken()
		}
	}
	if in.TargetCount <= 0 {
		if existing != nil && existing.TargetCount > 0 {
			in.TargetCount = existing.TargetCount
		} else {
			in.TargetCount = 1
		}
	}
	return nil
}

func (s *Service) Create(ctx context.Context, in MissionInput) (*domain.Mission, error) {
	if strings.TrimSpace(in.Title) == "" || in.HourWeight <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if err := s.normalize(&in, nil); err != nil {
		return nil, err
	}
	m := &domain.Mission{
		ID:            uuid.New(),
		Title:         strings.TrimSpace(in.Title),
		Description:   strings.TrimSpace(in.Description),
		Kind:          in.Kind,
		HourWeight:    in.HourWeight,
		DeadlineHours: in.DeadlineHours,
		WebhookEvent:  in.WebhookEvent,
		TargetCount:   in.TargetCount,
		VerifyMode:    in.VerifyMode,
		VerifyURL:     in.VerifyURL,
		VerifyToken:   in.VerifyToken,
		Status:        domain.MissionActive,
		CreatedAt:     s.clock.Now(),
	}
	if in.Status != "" {
		m.Status = in.Status
	}
	return m, s.missions.Create(ctx, m)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in MissionInput) (*domain.Mission, error) {
	m, err := s.missions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.normalize(&in, m); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Title) != "" {
		m.Title = strings.TrimSpace(in.Title)
	}
	m.Description = strings.TrimSpace(in.Description)
	m.Kind = in.Kind
	if in.HourWeight > 0 {
		m.HourWeight = in.HourWeight
	}
	m.DeadlineHours = in.DeadlineHours
	m.WebhookEvent = in.WebhookEvent
	m.TargetCount = in.TargetCount
	m.VerifyMode = in.VerifyMode
	m.VerifyURL = in.VerifyURL
	m.VerifyToken = in.VerifyToken
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
	return s.ListProgressForVolunteer(ctx, v.ID)
}

func (s *Service) ListProgressForVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]domain.MissionProgress, error) {
	items, err := s.missions.ListProgressByVolunteer(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.MissionProgress{}
	}
	return items, nil
}

func (s *Service) Verify(ctx context.Context, userID, missionID uuid.UUID) (*domain.MissionProgress, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.verifyVolunteer(ctx, v, missionID)
}

func (s *Service) VerifyKind(ctx context.Context, userID uuid.UUID, kind domain.MissionKind) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return
	}
	items, err := s.missions.List(ctx, true)
	if err != nil {
		return
	}
	for _, m := range items {
		if m.Kind != kind || m.VerifyMode != domain.VerifyInternal {
			continue
		}
		_, _ = s.verifyVolunteer(ctx, v, m.ID)
	}
}

func (s *Service) AwardInbound(ctx context.Context, token, event, volunteerID, phone string, increment int) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.ErrUnauthorized
	}
	m, err := s.missions.GetByVerifyToken(ctx, token)
	if err != nil {
		return domain.ErrUnauthorized
	}
	if event != "" && m.WebhookEvent != "" && m.WebhookEvent != event {
		return domain.Invalid("رویداد با این ماموریت هم‌خوانی ندارد")
	}
	var v *domain.Volunteer
	if volunteerID != "" {
		id, perr := uuid.Parse(volunteerID)
		if perr != nil {
			return domain.ErrInvalidInput
		}
		v, err = s.volunteers.GetByID(ctx, id)
	} else if strings.TrimSpace(phone) != "" {
		v, err = s.volunteers.GetByPhone(ctx, strings.TrimSpace(phone))
	} else {
		return domain.Invalid("شناسه داوطلب یا شماره موبایل لازم است")
	}
	if err != nil {
		return err
	}
	if increment <= 0 {
		increment = 1
	}
	p, err := s.loadProgress(ctx, v, m)
	if err != nil {
		return err
	}
	if p.Status == domain.MissionCompleted {
		return nil
	}
	_, err = s.applyCount(ctx, v, m, p, p.Progress+increment, "")
	return err
}

func (s *Service) verifyVolunteer(ctx context.Context, v *domain.Volunteer, missionID uuid.UUID) (*domain.MissionProgress, error) {
	m, err := s.missions.GetByID(ctx, missionID)
	if err != nil {
		return nil, err
	}
	p, err := s.loadProgress(ctx, v, m)
	if err != nil {
		return nil, err
	}
	if p.Status == domain.MissionCompleted {
		p.Mission = publicProgressMission(m)
		return p, nil
	}
	count, reason, err := s.verifiedCount(ctx, v, m)
	if err != nil {
		return p, err
	}
	p, err = s.applyCount(ctx, v, m, p, count, reason)
	if err != nil {
		return p, err
	}
	if p.Status != domain.MissionCompleted {
		if reason == "" {
			reason = "انجام این ماموریت هنوز توسط سرویس مربوطه تأیید نشده است"
		}
		return p, domain.Invalid(reason)
	}
	return p, nil
}

func (s *Service) loadProgress(ctx context.Context, v *domain.Volunteer, m *domain.Mission) (*domain.MissionProgress, error) {
	p, err := s.missions.GetProgress(ctx, m.ID, v.ID)
	now := s.clock.Now()
	if err == domain.ErrNotFound {
		p = &domain.MissionProgress{
			ID:          uuid.New(),
			MissionID:   m.ID,
			VolunteerID: v.ID,
			Status:      domain.MissionInProgress,
			StartedAt:   now,
		}
		if m.DeadlineHours != nil {
			due := now.Add(time.Duration(*m.DeadlineHours) * time.Hour)
			p.DueAt = &due
		}
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	if p.DueAt != nil && now.After(*p.DueAt) && p.Status != domain.MissionCompleted {
		p.Status = domain.MissionExpired
		_ = s.missions.UpsertProgress(ctx, p)
		return p, domain.ErrMissionExpired
	}
	return p, nil
}

func (s *Service) applyCount(ctx context.Context, v *domain.Volunteer, m *domain.Mission, p *domain.MissionProgress, count int, _ string) (*domain.MissionProgress, error) {
	if count < 0 {
		count = 0
	}
	if count < p.Progress {
		count = p.Progress
	}
	p.Progress = count
	now := s.clock.Now()
	if p.Progress >= m.TargetCount {
		p.Progress = m.TargetCount
		p.Status = domain.MissionCompleted
		p.CompletedAt = &now
		scoring.UpdateVolunteerTotals(v, 5, m.HourWeight)
		v.UpdatedAt = now
		if err := s.volunteers.Update(ctx, v); err != nil {
			return nil, err
		}
		if s.notify != nil {
			_ = s.notify.Notify(ctx, v.UserID, "ماموریت تکمیل شد", "امتیاز و ساعات داوطلبی ماموریت «"+m.Title+"» به حساب شما افزوده شد.")
		}
	}
	if err := s.missions.UpsertProgress(ctx, p); err != nil {
		return nil, err
	}
	p.Mission = publicProgressMission(m)
	return p, nil
}

func publicProgressMission(m *domain.Mission) *domain.Mission {
	if m == nil {
		return nil
	}
	cp := m.Public()
	return &cp
}
