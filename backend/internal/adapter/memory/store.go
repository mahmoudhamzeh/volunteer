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
	certs       map[uuid.UUID]*domain.Certificate
	certReqs    map[uuid.UUID]*domain.CertificateRequest
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
		certs:       map[uuid.UUID]*domain.Certificate{},
		certReqs:    map[uuid.UUID]*domain.CertificateRequest{},
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
func (a VolunteerAdapter) ReplaceSkills(_ context.Context, volunteerID uuid.UUID, skillIDs []uuid.UUID) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	v, ok := a.S.volunteers[volunteerID]
	if !ok {
		return domain.ErrNotFound
	}
	skills := make([]domain.VolunteerSkill, 0, len(skillIDs))
	for _, id := range skillIDs {
		if id == uuid.Nil {
			continue
		}
		skills = append(skills, domain.VolunteerSkill{SkillID: id})
	}
	v.Skills = skills
	return nil
}
func (a VolunteerAdapter) ListVolunteerSkills(_ context.Context, volunteerID uuid.UUID) ([]domain.VolunteerSkill, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	v, ok := a.S.volunteers[volunteerID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return append([]domain.VolunteerSkill{}, v.Skills...), nil
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
func (a TaskAdapter) Delete(_ context.Context, id uuid.UUID) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	delete(a.S.tasks, id)
	return nil
}
func (a TaskAdapter) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return a.S.GetTask(ctx, id)
}
func (a TaskAdapter) List(_ context.Context, f domain.TaskFilter) ([]domain.Task, int, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var out []domain.Task
	for _, t := range a.S.tasks {
		if f.Status != "" && t.Status != f.Status {
			continue
		}
		if f.Kind != "" && t.Kind != f.Kind {
			continue
		}
		if f.ExcludeKind != "" && t.Kind == f.ExcludeKind {
			continue
		}
		if f.SeriesID != uuid.Nil && t.SeriesID != f.SeriesID {
			continue
		}
		cp := *t
		if t.Slots != nil {
			cp.Slots = append([]domain.TaskSlot{}, t.Slots...)
		}
		out = append(out, cp)
	}
	return out, len(out), nil
}
func (a TaskAdapter) CloseExpired(_ context.Context, now time.Time) (int64, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var n int64
	for _, t := range a.S.tasks {
		if t.Status == domain.TaskOpen && !t.EndsAt.IsZero() && t.EndsAt.Before(now) {
			t.Status = domain.TaskClosed
			t.UpdatedAt = now
			n++
		}
	}
	return n, nil
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
func (a TaskAdapter) ListAssignments(_ context.Context, f domain.AssignmentFilter) ([]domain.Assignment, int, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var out []domain.Assignment
	for _, x := range a.S.assignments {
		if f.VolunteerID != uuid.Nil && x.VolunteerID != f.VolunteerID {
			continue
		}
		if f.TaskID != uuid.Nil && x.TaskID != f.TaskID {
			continue
		}
		if f.SeriesID != uuid.Nil {
			t := a.S.tasks[x.TaskID]
			if t == nil {
				continue
			}
			if t.SeriesID != f.SeriesID && x.TaskID != f.SeriesID {
				continue
			}
		}
		if f.Status != "" && x.Status != f.Status {
			continue
		}
		cp := *x
		if t, ok := a.S.tasks[x.TaskID]; ok {
			tc := *t
			cp.Task = &tc
		}
		if v, ok := a.S.volunteers[x.VolunteerID]; ok {
			vc := *v
			cp.Volunteer = &vc
		}
		out = append(out, cp)
	}
	return out, len(out), nil
}

type CertAdapter struct{ S *Store }

func (a CertAdapter) Create(_ context.Context, c *domain.Certificate) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	cp := *c
	a.S.certs[c.ID] = &cp
	return nil
}

func (a CertAdapter) GetByVerificationCode(_ context.Context, code uuid.UUID) (*domain.Certificate, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	for _, c := range a.S.certs {
		if c.VerificationCode == code {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (a CertAdapter) GetByAssignment(_ context.Context, assignmentID uuid.UUID) (*domain.Certificate, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	for _, c := range a.S.certs {
		if c.AssignmentID != nil && *c.AssignmentID == assignmentID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (a CertAdapter) ListByVolunteer(_ context.Context, volunteerID uuid.UUID) ([]domain.Certificate, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var out []domain.Certificate
	for _, c := range a.S.certs {
		if c.VolunteerID == volunteerID {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (a CertAdapter) ExistsForAssignment(ctx context.Context, assignmentID uuid.UUID) (bool, error) {
	_, err := a.GetByAssignment(ctx, assignmentID)
	if err == domain.ErrNotFound {
		return false, nil
	}
	return err == nil, err
}

func (a CertAdapter) CreateRequest(_ context.Context, req *domain.CertificateRequest) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	cp := *req
	a.S.certReqs[req.ID] = &cp
	return nil
}

func (a CertAdapter) GetRequest(_ context.Context, id uuid.UUID) (*domain.CertificateRequest, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	req, ok := a.S.certReqs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *req
	return &cp, nil
}

func (a CertAdapter) UpdateRequest(_ context.Context, req *domain.CertificateRequest) error {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	if _, ok := a.S.certReqs[req.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *req
	a.S.certReqs[req.ID] = &cp
	return nil
}

func (a CertAdapter) ListRequests(_ context.Context, status domain.CertificateRequestStatus) ([]domain.CertificateRequest, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var out []domain.CertificateRequest
	for _, req := range a.S.certReqs {
		if status != "" && req.Status != status {
			continue
		}
		out = append(out, a.hydrateCertReq(*req))
	}
	return out, nil
}

func (a CertAdapter) ListRequestsByVolunteer(_ context.Context, volunteerID uuid.UUID) ([]domain.CertificateRequest, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	var out []domain.CertificateRequest
	for _, req := range a.S.certReqs {
		if req.VolunteerID == volunteerID {
			out = append(out, a.hydrateCertReq(*req))
		}
	}
	return out, nil
}

func (a CertAdapter) HasPendingRequest(_ context.Context, volunteerID uuid.UUID, kind domain.CertificateKind, assignmentID *uuid.UUID) (bool, error) {
	a.S.mu.Lock()
	defer a.S.mu.Unlock()
	for _, req := range a.S.certReqs {
		if req.VolunteerID != volunteerID || req.Kind != kind || !req.Status.IsOpen() {
			continue
		}
		if assignmentID == nil && req.AssignmentID == nil {
			return true, nil
		}
		if assignmentID != nil && req.AssignmentID != nil && *assignmentID == *req.AssignmentID {
			return true, nil
		}
	}
	return false, nil
}

func (a CertAdapter) hydrateCertReq(req domain.CertificateRequest) domain.CertificateRequest {
	if v, ok := a.S.volunteers[req.VolunteerID]; ok {
		req.VolunteerName = v.FullName
	}
	if req.AssignmentID != nil {
		if asg, ok := a.S.assignments[*req.AssignmentID]; ok {
			if t, ok := a.S.tasks[asg.TaskID]; ok {
				req.AssignmentTitle = t.Title
			}
		}
	}
	return req
}
