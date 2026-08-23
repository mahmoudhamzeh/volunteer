package authuc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/memory"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type userMem struct {
	byID    map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func newUserMem() *userMem {
	return &userMem{byID: map[uuid.UUID]*domain.User{}, byEmail: map[string]*domain.User{}}
}

func (m *userMem) Create(_ context.Context, u *domain.User) error {
	if _, ok := m.byEmail[u.Email]; ok {
		return domain.ErrConflict
	}
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	return nil
}
func (m *userMem) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (m *userMem) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (m *userMem) GetByExternalID(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func TestRegisterIgnoresRequestedAdminRole(t *testing.T) {
	users := newUserMem()
	store := memory.New()
	svc := New(users, memory.VolunteerAdapter{S: store}, "unit-test-secret-unit-test-secret", time.Hour)
	u, token, err := svc.Register(context.Background(), "a@mahak.ir", "Password1", "علی", domain.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != domain.RoleVolunteer {
		t.Fatalf("role=%s", u.Role)
	}
	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Role != domain.RoleVolunteer {
		t.Fatalf("token role=%s", claims.Role)
	}
}

func TestParseRejectsWrongAlgorithm(t *testing.T) {
	svc := New(newUserMem(), memory.VolunteerAdapter{S: memory.New()}, "unit-test-secret-unit-test-secret", time.Hour)
	if _, err := svc.Parse("not-a-token"); err != domain.ErrUnauthorized {
		t.Fatalf("err=%v", err)
	}
}
