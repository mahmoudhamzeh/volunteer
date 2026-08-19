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
	documents   map[uuid.UUID]*domain.Document
	events      []domain.VolunteerEvent
}

func New() *Store {
	return &Store{
		users:       map[uuid.UUID]*domain.User{},
		volunteers:  map[uuid.UUID]*domain.Volunteer{},
		byUser:      map[uuid.UUID]uuid.UUID{},
		tasks:       map[uuid.UUID]*domain.Task{},
		assignments: map[uuid.UUID]*domain.Assignment{},
		documents:   map[uuid.UUID]*domain.Document{},
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

func (s *Store) ApplySeat(_ context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[taskID]; !ok {
		return nil, domain.ErrNotFound
	}
	for _, a := range s.assignments {
		if a.TaskID == taskID && a.VolunteerID == volunteerID {
			if a.Status.BlocksReapply() {
				return nil, domain.ErrAlreadyAssigned
			}
			a.Status = domain.AssignmentRequested
			a.AdminComment = ""
			cp := *a
			return &cp, nil
		}
	}
	a := &domain.Assignment{
		ID:          uuid.New(),
		TaskID:      taskID,
		VolunteerID: volunteerID,
		Status:      domain.AssignmentRequested,
		CreatedAt:   time.Now().UTC(),
	}
	s.assignments[a.ID] = a
	cp := *a
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
func (a VolunteerAdapter) GetByPhone(_ context.Context, phone string) (*domain.Volunteer, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	for _, v := range a.S.volunteers {
		if v.Phone == phone && phone != "" {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
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
func (a VolunteerAdapter) AddDocument(_ context.Context, d *domain.Document) error {
	if d == nil {
		return nil
	}
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	cp := *d
	a.S.documents[d.ID] = &cp
	return nil
}
func (a VolunteerAdapter) ListDocuments(_ context.Context, volunteerID uuid.UUID) ([]domain.Document, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var out []domain.Document
	for _, d := range a.S.documents {
		if d.VolunteerID == volunteerID {
			out = append(out, *d)
		}
	}
	return out, nil
}
func (a VolunteerAdapter) GetDocument(_ context.Context, id uuid.UUID) (*domain.Document, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	d, ok := a.S.documents[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *d
	return &cp, nil
}
func (a VolunteerAdapter) DeleteDocument(_ context.Context, id uuid.UUID) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	if _, ok := a.S.documents[id]; !ok {
		return domain.ErrNotFound
	}
	delete(a.S.documents, id)
	return nil
}
func (a VolunteerAdapter) AddEvent(_ context.Context, e *domain.VolunteerEvent) error {
	if e == nil {
		return nil
	}
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	cp := *e
	a.S.events = append(a.S.events, cp)
	return nil
}
func (a VolunteerAdapter) ListEvents(_ context.Context, volunteerID uuid.UUID, limit int) ([]domain.VolunteerEvent, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var out []domain.VolunteerEvent
	for i := len(a.S.events) - 1; i >= 0; i-- {
		if a.S.events[i].VolunteerID == volunteerID {
			out = append(out, a.S.events[i])
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	if out == nil {
		out = []domain.VolunteerEvent{}
	}
	return out, nil
}
func (a VolunteerAdapter) ReplaceSkills(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}
func (a VolunteerAdapter) ListVolunteerSkills(context.Context, uuid.UUID) ([]domain.VolunteerSkill, error) {
	return []domain.VolunteerSkill{}, nil
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
func (a TaskAdapter) ApplySeat(ctx context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	return a.S.ApplySeat(ctx, taskID, volunteerID)
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
func (a TaskAdapter) GetAssignmentByTaskVolunteer(_ context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	for _, x := range a.S.assignments {
		if x.TaskID == taskID && x.VolunteerID == volunteerID {
			cp := *x
			return &cp, nil
		}
	}
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
