package authuc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type fakeUsers struct {
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
	byExt   map[string]*domain.User
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byEmail: map[string]*domain.User{}, byID: map[uuid.UUID]*domain.User{}, byExt: map[string]*domain.User{}}
}

func (f *fakeUsers) Create(_ context.Context, u *domain.User) error {
	if _, ok := f.byEmail[u.Email]; ok {
		return domain.ErrConflict
	}
	cp := *u
	f.byEmail[u.Email] = &cp
	f.byID[u.ID] = &cp
	if u.ExternalUserID != "" {
		f.byExt[u.ExternalUserID] = &cp
	}
	return nil
}
func (f *fakeUsers) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (f *fakeUsers) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (f *fakeUsers) GetByPhone(_ context.Context, _ string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeUsers) GetByExternalID(_ context.Context, externalID string) (*domain.User, error) {
	u, ok := f.byExt[externalID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

type fakeVols struct {
	byUser map[uuid.UUID]*domain.Volunteer
}

func (f *fakeVols) Create(_ context.Context, v *domain.Volunteer) error {
	if f.byUser == nil {
		f.byUser = map[uuid.UUID]*domain.Volunteer{}
	}
	cp := *v
	f.byUser[v.UserID] = &cp
	return nil
}
func (f *fakeVols) Update(context.Context, *domain.Volunteer) error { return nil }
func (f *fakeVols) GetByID(context.Context, uuid.UUID) (*domain.Volunteer, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeVols) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	v, ok := f.byUser[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *v
	return &cp, nil
}
func (f *fakeVols) GetByPhone(context.Context, string) (*domain.Volunteer, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeVols) List(context.Context, domain.VolunteerFilter) ([]domain.Volunteer, int, error) {
	return nil, 0, nil
}
func (f *fakeVols) ReplaceAvailability(context.Context, uuid.UUID, []domain.AvailabilitySlot) error {
	return nil
}
func (f *fakeVols) ListAvailability(context.Context, uuid.UUID) ([]domain.AvailabilitySlot, error) {
	return nil, nil
}
func (f *fakeVols) AddDocument(context.Context, *domain.Document) error { return nil }
func (f *fakeVols) ListDocuments(context.Context, uuid.UUID) ([]domain.Document, error) {
	return nil, nil
}
func (f *fakeVols) GetDocument(context.Context, uuid.UUID) (*domain.Document, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeVols) DeleteDocument(context.Context, uuid.UUID) error { return nil }
func (f *fakeVols) AddEvent(context.Context, *domain.VolunteerEvent) error {
	return nil
}
func (f *fakeVols) ListEvents(context.Context, uuid.UUID, int) ([]domain.VolunteerEvent, error) {
	return nil, nil
}
func (f *fakeVols) ReplaceSkills(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (f *fakeVols) ListVolunteerSkills(context.Context, uuid.UUID) ([]domain.VolunteerSkill, error) {
	return nil, nil
}

func TestRegisterIgnoresAdminRole(t *testing.T) {
	users := newFakeUsers()
	vols := &fakeVols{}
	svc := New(users, vols, "test-secret", time.Hour)
	u, token, err := svc.Register(context.Background(), "a@mahak.ir", "Password1", "Ali", domain.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != domain.RoleVolunteer {
		t.Fatalf("role=%s want volunteer", u.Role)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Role != domain.RoleVolunteer {
		t.Fatalf("claims role=%s", claims.Role)
	}
}

func TestLoginAndParse(t *testing.T) {
	users := newFakeUsers()
	vols := &fakeVols{}
	svc := New(users, vols, "test-secret", time.Hour)
	_, _, err := svc.Register(context.Background(), "v@mahak.ir", "Password1", "Sara", "")
	if err != nil {
		t.Fatal(err)
	}
	u, token, err := svc.Login(context.Background(), "v@mahak.ir", "Password1")
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "v@mahak.ir" || token == "" {
		t.Fatalf("unexpected login: %+v", u)
	}
	if _, _, err := svc.Login(context.Background(), "v@mahak.ir", "wrongpass"); err != domain.ErrUnauthorized {
		t.Fatalf("want unauthorized, got %v", err)
	}
}
