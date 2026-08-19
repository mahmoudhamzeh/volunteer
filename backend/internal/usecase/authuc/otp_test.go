package authuc_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/authuc"
)

type memUsers struct {
	mu      sync.Mutex
	byID    map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
	byPhone map[string]*domain.User
}

func newMemUsers() *memUsers {
	return &memUsers{
		byID:    map[uuid.UUID]*domain.User{},
		byEmail: map[string]*domain.User{},
		byPhone: map[string]*domain.User{},
	}
}

func (m *memUsers) Create(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	if u.Phone != "" {
		m.byPhone[u.Phone] = &cp
	}
	return nil
}
func (m *memUsers) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (m *memUsers) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (m *memUsers) GetByPhone(_ context.Context, phone string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byPhone[phone]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (m *memUsers) GetByExternalID(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

type memVols struct {
	mu      sync.Mutex
	byUser  map[uuid.UUID]*domain.Volunteer
	byPhone map[string]*domain.Volunteer
}

func newMemVols() *memVols {
	return &memVols{byUser: map[uuid.UUID]*domain.Volunteer{}, byPhone: map[string]*domain.Volunteer{}}
}

func (m *memVols) Create(_ context.Context, v *domain.Volunteer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *v
	m.byUser[v.UserID] = &cp
	if v.Phone != "" {
		m.byPhone[v.Phone] = &cp
	}
	return nil
}
func (m *memVols) Update(context.Context, *domain.Volunteer) error { return nil }
func (m *memVols) GetByID(context.Context, uuid.UUID) (*domain.Volunteer, error) {
	return nil, domain.ErrNotFound
}
func (m *memVols) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.byUser[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *v
	return &cp, nil
}
func (m *memVols) GetByPhone(_ context.Context, phone string) (*domain.Volunteer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.byPhone[phone]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *v
	return &cp, nil
}
func (m *memVols) List(context.Context, domain.VolunteerFilter) ([]domain.Volunteer, int, error) {
	return nil, 0, nil
}
func (m *memVols) ReplaceAvailability(context.Context, uuid.UUID, []domain.AvailabilitySlot) error {
	return nil
}
func (m *memVols) ListAvailability(context.Context, uuid.UUID) ([]domain.AvailabilitySlot, error) {
	return nil, nil
}
func (m *memVols) AddDocument(context.Context, *domain.Document) error { return nil }
func (m *memVols) ListDocuments(context.Context, uuid.UUID) ([]domain.Document, error) {
	return nil, nil
}
func (m *memVols) GetDocument(context.Context, uuid.UUID) (*domain.Document, error) {
	return nil, domain.ErrNotFound
}
func (m *memVols) DeleteDocument(context.Context, uuid.UUID) error { return nil }
func (m *memVols) AddEvent(context.Context, *domain.VolunteerEvent) error {
	return nil
}
func (m *memVols) ListEvents(context.Context, uuid.UUID, int) ([]domain.VolunteerEvent, error) {
	return nil, nil
}
func (m *memVols) ReplaceSkills(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (m *memVols) ListVolunteerSkills(context.Context, uuid.UUID) ([]domain.VolunteerSkill, error) {
	return []domain.VolunteerSkill{}, nil
}

func TestNormalizeMobile(t *testing.T) {
	got, err := authuc.NormalizeMobile("۰۹۱۲۱۲۳۴۵۶۷")
	if err != nil || got != "09121234567" {
		t.Fatalf("got %q err=%v", got, err)
	}
	got, err = authuc.NormalizeMobile("+98 912 123 4567")
	if err != nil || got != "09121234567" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := authuc.NormalizeMobile("02188990011"); err == nil {
		t.Fatal("landline should fail")
	}
}

func TestSendAndVerifyOTPCreatesVolunteer(t *testing.T) {
	svc := authuc.New(newMemUsers(), newMemVols(), "test-secret", time.Hour)
	ctx := context.Background()
	sent, err := svc.SendOTP(ctx, "09121234567")
	if err != nil {
		t.Fatal(err)
	}
	if sent.Phone != "09121234567" || sent.DevCode == "" || !sent.IsNew {
		t.Fatalf("%+v", sent)
	}
	if _, err := svc.SendOTP(ctx, "09121234567"); err == nil {
		t.Fatal("cooldown should block resend")
	}
	u, token, isNew, err := svc.VerifyOTP(ctx, "0912 123 4567", sent.DevCode, "سارا محمدی")
	if err != nil {
		t.Fatal(err)
	}
	if !isNew || token == "" || u.Phone != "09121234567" || u.Role != domain.RoleVolunteer {
		t.Fatalf("user=%+v isNew=%v", u, isNew)
	}
	again, err := svc.SendOTP(ctx, "09121234567")
	if err != nil {
		t.Fatal(err)
	}
	if again.IsNew {
		t.Fatal("existing phone should not be new")
	}
	u2, _, isNew2, err := svc.VerifyOTP(ctx, "09121234567", again.DevCode, "")
	if err != nil {
		t.Fatal(err)
	}
	if isNew2 || u2.ID != u.ID {
		t.Fatalf("want same user, isNew=%v", isNew2)
	}
}

func TestVerifyOTPRejectsWrongCode(t *testing.T) {
	svc := authuc.New(newMemUsers(), newMemVols(), "test-secret", time.Hour)
	ctx := context.Background()
	if _, err := svc.SendOTP(ctx, "09351234567"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := svc.VerifyOTP(ctx, "09351234567", "11111", ""); err == nil {
		t.Fatal("wrong code should fail")
	}
}
