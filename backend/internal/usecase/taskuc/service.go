package taskuc

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/scoring"
)

type Service struct {
	tasks      domain.TaskRepository
	volunteers domain.VolunteerRepository
	certs      domain.CertificateRepository
	locker     domain.Locker
	notify     domain.Notifier
	clock      domain.Clock
}

func New(tasks domain.TaskRepository, volunteers domain.VolunteerRepository, certs domain.CertificateRepository, locker domain.Locker, notify domain.Notifier, clock domain.Clock) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	return &Service{tasks: tasks, volunteers: volunteers, certs: certs, locker: locker, notify: notify, clock: clock}
}

type TaskInput struct {
	Title             string
	Description       string
	Location          string
	StartsAt          time.Time
	EndsAt            time.Time
	Capacity          int
	HourWeight        float64
	RequiredSkills    []string
	MinScore          float64
	RequiredEducation string
	Status            domain.TaskStatus
}

func (s *Service) Create(ctx context.Context, actor uuid.UUID, in TaskInput) (*domain.Task, error) {
	if err := validateTask(in); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	t := &domain.Task{
		ID:                uuid.New(),
		Title:             strings.TrimSpace(in.Title),
		Description:       strings.TrimSpace(in.Description),
		Location:          strings.TrimSpace(in.Location),
		StartsAt:          in.StartsAt.UTC(),
		EndsAt:            in.EndsAt.UTC(),
		Capacity:          in.Capacity,
		HourWeight:        in.HourWeight,
		RequiredSkills:    domain.ParseSkillCategories(in.RequiredSkills),
		MinScore:          in.MinScore,
		RequiredEducation: strings.TrimSpace(in.RequiredEducation),
		Status:            domain.TaskOpen,
		CreatedBy:         actor,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if in.Status != "" {
		t.Status = in.Status
	}
	return t, s.tasks.Create(ctx, t)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in TaskInput) (*domain.Task, error) {
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := validateTask(in); err != nil {
		return nil, err
	}
	t.Title = strings.TrimSpace(in.Title)
	t.Description = strings.TrimSpace(in.Description)
	t.Location = strings.TrimSpace(in.Location)
	t.StartsAt = in.StartsAt.UTC()
	t.EndsAt = in.EndsAt.UTC()
	t.Capacity = in.Capacity
	t.HourWeight = in.HourWeight
	t.RequiredSkills = domain.ParseSkillCategories(in.RequiredSkills)
	t.MinScore = in.MinScore
	t.RequiredEducation = strings.TrimSpace(in.RequiredEducation)
	if in.Status != "" {
		t.Status = in.Status
	}
	t.UpdatedAt = s.clock.Now()
	return t, s.tasks.Update(ctx, t)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.tasks.Delete(ctx, id)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return s.tasks.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.tasks.List(ctx, f)
}

func (s *Service) ListEligible(ctx context.Context, userID uuid.UUID, f domain.TaskFilter) ([]domain.Task, int, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if !v.Status.CanViewTasks() {
		return nil, 0, domain.ErrNotApproved
	}
	f.Status = domain.TaskOpen
	f.Upcoming = true
	tasks, total, err := s.tasks.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.Task, 0, len(tasks))
	for _, t := range tasks {
		if scoring.EligibleForTask(*v, t) == nil {
			out = append(out, t)
		}
	}
	return out, total, nil
}

func (s *Service) Accept(ctx context.Context, userID, taskID uuid.UUID) (*domain.Assignment, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !t.IsOpen() {
		return nil, domain.ErrNotEligible
	}
	if err := scoring.EligibleForTask(*v, *t); err != nil {
		return nil, err
	}
	unlock := func() {}
	if s.locker != nil {
		unlockFn, err := s.locker.Lock(ctx, "task:"+taskID.String(), 8*time.Second)
		if err != nil {
			return nil, err
		}
		unlock = unlockFn
	}
	defer unlock()
	asg, err := s.tasks.ReserveSeat(ctx, taskID, v.ID)
	if err != nil {
		return nil, err
	}
	asg.Task = t
	asg.Volunteer = v
	return asg, nil
}

func (s *Service) ConfirmAttendance(ctx context.Context, assignmentID uuid.UUID) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AssignmentReserved {
		return nil, domain.ErrInvalidTransition
	}
	now := s.clock.Now()
	a.Status = domain.AssignmentAttended
	a.AttendedAt = &now
	return a, s.tasks.UpdateAssignment(ctx, a)
}

func (s *Service) Complete(ctx context.Context, assignmentID uuid.UUID, discipline, expertise, ethics int, comment string) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AssignmentAttended && a.Status != domain.AssignmentReserved {
		return nil, domain.ErrInvalidTransition
	}
	score, err := scoring.CompositeScore(discipline, expertise, ethics)
	if err != nil {
		return nil, err
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	v, err := s.volunteers.GetByID(ctx, a.VolunteerID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	a.AdminDiscipline = &discipline
	a.AdminExpertise = &expertise
	a.AdminEthics = &ethics
	a.AdminComment = comment
	a.CompositeScore = &score
	a.HoursAwarded = t.HourWeight
	a.Status = domain.AssignmentCompleted
	a.CompletedAt = &now
	if a.AttendedAt == nil {
		a.AttendedAt = &now
	}
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	scoring.UpdateVolunteerTotals(v, score, t.HourWeight)
	v.UpdatedAt = now
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	if s.notify != nil {
		_ = s.notify.Notify(ctx, v.UserID, "تسک تکمیل شد", "امتیاز شما ثبت شد. در صورت تایید نهایی، گواهی صادر می‌شود.")
	}
	a.Task = t
	a.Volunteer = v
	return a, nil
}

func (s *Service) RateByVolunteer(ctx context.Context, userID, assignmentID uuid.UUID, rating int, comment string) (*domain.Assignment, error) {
	if rating < 1 || rating > 5 {
		return nil, domain.ErrInvalidInput
	}
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.VolunteerID != v.ID {
		return nil, domain.ErrForbidden
	}
	if a.Status != domain.AssignmentCompleted && a.Status != domain.AssignmentAttended {
		return nil, domain.ErrInvalidTransition
	}
	a.VolunteerRating = &rating
	a.VolunteerComment = comment
	return a, s.tasks.UpdateAssignment(ctx, a)
}

func (s *Service) Cancel(ctx context.Context, assignmentID uuid.UUID, byAdmin bool) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AssignmentReserved {
		return nil, domain.ErrInvalidTransition
	}
	if byAdmin {
		a.Status = domain.AssignmentRejected
	} else {
		a.Status = domain.AssignmentCancelled
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if t.ReservedCount > 0 {
		t.ReservedCount--
		_ = s.tasks.Update(ctx, t)
	}
	return a, s.tasks.UpdateAssignment(ctx, a)
}

func (s *Service) ListAssignments(ctx context.Context, f domain.AssignmentFilter) ([]domain.Assignment, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.tasks.ListAssignments(ctx, f)
}

func (s *Service) MyAssignments(ctx context.Context, userID uuid.UUID) ([]domain.Assignment, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, _, err := s.tasks.ListAssignments(ctx, domain.AssignmentFilter{VolunteerID: v.ID, Limit: 100})
	return items, err
}

func validateTask(in TaskInput) error {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Description) == "" {
		return domain.ErrInvalidInput
	}
	if in.Capacity < 1 || in.HourWeight <= 0 {
		return domain.ErrInvalidInput
	}
	if in.EndsAt.Before(in.StartsAt) {
		return domain.ErrInvalidInput
	}
	return nil
}
