package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/memory"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/authuc"
)

type simpleUsers struct {
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
}

func (m *simpleUsers) Create(_ context.Context, u *domain.User) error {
	if m.byID == nil {
		m.byID = map[uuid.UUID]*domain.User{}
	}
	if _, ok := m.byEmail[u.Email]; ok {
		return domain.ErrConflict
	}
	cp := *u
	m.byEmail[u.Email] = &cp
	m.byID[u.ID] = &cp
	return nil
}

func (m *simpleUsers) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *simpleUsers) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *simpleUsers) GetByExternalID(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func TestHealthz(t *testing.T) {
	r := NewRouter(Deps{Ready: func() map[string]string { return map[string]string{"status": "ready"} }})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestRegisterCannotBecomeAdmin(t *testing.T) {
	store := memory.New()
	users := &simpleUsers{byEmail: map[string]*domain.User{}}
	auth := authuc.New(users, memory.VolunteerAdapter{S: store}, "unit-test-secret-unit-test-secret", time.Hour)
	r := NewRouter(Deps{Auth: auth, Users: users})
	body, _ := json.Marshal(map[string]string{
		"email": "boss@mahak.ir", "password": "Password1", "full_name": "مدیر", "role": "admin",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		User struct {
			Role string `json:"role"`
		} `json:"user"`
	}
	_ = json.NewDecoder(w.Body).Decode(&out)
	if out.User.Role != "volunteer" {
		t.Fatalf("role=%s", out.User.Role)
	}
}

func TestReadyzNotReady(t *testing.T) {
	r := NewRouter(Deps{Ready: func() map[string]string { return map[string]string{"status": "not_ready", "postgres": "error"} }})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", w.Code)
	}
}
