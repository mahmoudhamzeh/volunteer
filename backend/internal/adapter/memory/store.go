package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

// Store is an in-memory implementation used by unit tests.
type Store struct {
	mu          sync.Mutex
	users       map[uuid.UUID]*domain.User
	volunteers  map[uuid.UUID]*domain.Volunteer
	byUser      map[uuid.UUID]uuid.UUID
	tasks       map[uuid.UUID]*domain.Task
	assignments map[uuid.UUID]*domain.Assignment
}

func New() *Store {
	return &Store{
		users:       map[uuid.UUID]*domain.User{},
		volunteers:  map[uuid.UUID]*domain.Volunteer{},
		byUser:      map[uuid.UUID]uuid.UUID{},
		tasks:       map[uuid.UUID]*domain.Task{},
		assignments: map[uuid.UUID]*domain.Assignment{},
	}
}

func (s *Store) CreateVolunteer(_ context.Context, v *domain.Volunteer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *v
	s.volunteers[v.ID] = &cp
	s.byUser[v.UserID] = v.ID
	return nil
}

func (s *Store) UpdateVolunteer(_ context.Context, v *domain.Volunteer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.volunteers[v.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *v
	s.volunteers[v.ID] = &cp
	return nil
}

func (s *Store) GetVolunteer(_ context.Context, id uuid.UUID) (*domain.Volunteer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.volunteers[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *v
	return &cp, nil
}

func (s *Store) GetVolunteerByUser(_ context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byUser[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *s.volunteers[id]
	return &cp, nil
}

func (s *Store) CreateTask(_ context.Context, t *domain.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *t
	s.tasks[t.ID] = &cp
	return nil
}

func (s *Store) GetTask(_ context.Context, id uuid.UUID) (*domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (s *Store) ReserveSeat(_ context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[taskID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if t.ReservedCount >= t.Capacity {
		return nil, domain.ErrCapacityFull
	}
	for _, a := range s.assignments {
		if a.TaskID == taskID && a.VolunteerID == volunteerID && a.Status != domain.AssignmentCancelled && a.Status != domain.AssignmentRejected {
			return nil, domain.ErrConflict
		}
	}
	t.ReservedCount++
	a := &domain.Assignment{
		ID:          uuid.New(),
		TaskID:      taskID,
		VolunteerID: volunteerID,
		Status:      domain.AssignmentReserved,
		CreatedAt:   time.Now().UTC(),
	}
	s.assignments[a.ID] = a
	cp := *a
	return &cp, nil
}

type VolunteerAdapter struct{ S *Store }

func (a VolunteerAdapter) Create(ctx context.Context, v *domain.Volunteer) error {
	return a.S.CreateVolunteer(ctx, v)
}
func (a VolunteerAdapter) Update(ctx context.Context, v *domain.Volunteer) error {
	return a.S.UpdateVolunteer(ctx, v)
}
func (a VolunteerAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volunteer, error) {
	return a.S.GetVolunteer(ctx, id)
}
func (a VolunteerAdapter) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	return a.S.GetVolunteerByUser(ctx, userID)
}
func (a VolunteerAdapter) List(context.Context, domain.VolunteerFilter) ([]domain.Volunteer, int, error) {
	return nil, 0, nil
}
func (a VolunteerAdapter) ReplaceAvailability(context.Context, uuid.UUID, []domain.AvailabilitySlot) error {
	return nil
}
func (a VolunteerAdapter) ListAvailability(context.Context, uuid.UUID) ([]domain.AvailabilitySlot, error) {
	return nil, nil
}
func (a VolunteerAdapter) AddDocument(context.Context, *domain.Document) error { return nil }
func (a VolunteerAdapter) ListDocuments(context.Context, uuid.UUID) ([]domain.Document, error) {
	return nil, nil
}
func (a VolunteerAdapter) GetDocument(context.Context, uuid.UUID) (*domain.Document, error) {
	return nil, domain.ErrNotFound
}
func (a VolunteerAdapter) AddCompletedWork(_ context.Context, volunteerID uuid.UUID, score, hours float64) (*domain.Volunteer, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	v, ok := a.S.volunteers[volunteerID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	completed := v.CompletedTasks
	if completed < 0 {
		completed = 0
	}
	total := float64(completed)*v.AverageScore + score
	v.CompletedTasks = completed + 1
	v.AverageScore = total / float64(v.CompletedTasks)
	v.TotalHours += hours
	cp := *v
	return &cp, nil
}

type TaskAdapter struct{ S *Store }

func (a TaskAdapter) Create(ctx context.Context, t *domain.Task) error { return a.S.CreateTask(ctx, t) }
func (a TaskAdapter) Update(_ context.Context, t *domain.Task) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	cp := *t
	a.S.tasks[t.ID] = &cp
	return nil
}
func (a TaskAdapter) Delete(context.Context, uuid.UUID) error { return nil }
func (a TaskAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return a.S.GetTask(ctx, id)
}
func (a TaskAdapter) List(context.Context, domain.TaskFilter) ([]domain.Task, int, error) {
	return nil, 0, nil
}
func (a TaskAdapter) ReserveSeat(ctx context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	return a.S.ReserveSeat(ctx, taskID, volunteerID)
}
func (a TaskAdapter) GetAssignment(_ context.Context, id uuid.UUID) (*domain.Assignment, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	x, ok := a.S.assignments[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *x
	return &cp, nil
}
func (a TaskAdapter) GetAssignmentByTaskVolunteer(context.Context, uuid.UUID, uuid.UUID) (*domain.Assignment, error) {
	return nil, domain.ErrNotFound
}
func (a TaskAdapter) UpdateAssignment(_ context.Context, asg *domain.Assignment) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	cp := *asg
	a.S.assignments[asg.ID] = &cp
	return nil
}
func (a TaskAdapter) ListAssignments(context.Context, domain.AssignmentFilter) ([]domain.Assignment, int, error) {
	return nil, 0, nil
}
func (a TaskAdapter) ListEligible(_ context.Context, v domain.Volunteer, f domain.TaskFilter) ([]domain.Task, int, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var all []domain.Task
	for _, t := range a.S.tasks {
		if t.Status != domain.TaskOpen || time.Now().After(t.EndsAt) {
			continue
		}
		if t.MinScore > 0 && v.CompletedTasks > 0 && v.AverageScore < t.MinScore {
			continue
		}
		if !v.HasAnySkill(t.RequiredSkills) {
			continue
		}
		if t.RequiredEducation != "" && v.EducationField != t.RequiredEducation {
			continue
		}
		taken := false
		for _, asg := range a.S.assignments {
			if asg.TaskID == t.ID && asg.VolunteerID == v.ID && asg.Status != domain.AssignmentCancelled && asg.Status != domain.AssignmentRejected {
				taken = true
				break
			}
		}
		if taken {
			continue
		}
		cp := *t
		all = append(all, cp)
	}
	total := len(all)
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	if f.Offset > total {
		return []domain.Task{}, total, nil
	}
	end := f.Offset + limit
	if end > total {
		end = total
	}
	return all[f.Offset:end], total, nil
}
func (a TaskAdapter) ReleaseSeat(_ context.Context, assignmentID uuid.UUID, next domain.AssignmentStatus) (*domain.Assignment, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	asg, ok := a.S.assignments[assignmentID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if asg.Status != domain.AssignmentReserved {
		return nil, domain.ErrInvalidTransition
	}
	asg.Status = next
	if t, ok := a.S.tasks[asg.TaskID]; ok && t.ReservedCount > 0 {
		t.ReservedCount--
	}
	cp := *asg
	return &cp, nil
}
