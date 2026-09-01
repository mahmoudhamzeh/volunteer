package missionuc_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/adapter/memory"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/missionuc"
)

type testClock struct{ t time.Time }

func (c testClock) Now() time.Time { return c.t }

type memMissions struct {
	items    map[uuid.UUID]*domain.Mission
	progress map[string]*domain.MissionProgress
}

func newMemMissions() *memMissions {
	return &memMissions{items: map[uuid.UUID]*domain.Mission{}, progress: map[string]*domain.MissionProgress{}}
}

func (m *memMissions) key(mid, vid uuid.UUID) string { return mid.String() + ":" + vid.String() }

func (m *memMissions) Create(_ context.Context, miss *domain.Mission) error {
	cp := *miss
	m.items[miss.ID] = &cp
	return nil
}
func (m *memMissions) Update(_ context.Context, miss *domain.Mission) error {
	if _, ok := m.items[miss.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *miss
	m.items[miss.ID] = &cp
	return nil
}
func (m *memMissions) GetByID(_ context.Context, id uuid.UUID) (*domain.Mission, error) {
	x, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *x
	return &cp, nil
}
func (m *memMissions) List(_ context.Context, activeOnly bool) ([]domain.Mission, error) {
	var out []domain.Mission
	for _, x := range m.items {
		if activeOnly && x.Status != domain.MissionActive {
			continue
		}
		out = append(out, *x)
	}
	return out, nil
}
func (m *memMissions) GetByWebhookEvent(_ context.Context, event string) ([]domain.Mission, error) {
	var out []domain.Mission
	for _, x := range m.items {
		if x.WebhookEvent == event && x.Status == domain.MissionActive {
			out = append(out, *x)
		}
	}
	return out, nil
}
func (m *memMissions) GetByVerifyToken(_ context.Context, token string) (*domain.Mission, error) {
	if token == "" {
		return nil, domain.ErrNotFound
	}
	for _, x := range m.items {
		if x.VerifyToken == token && x.Status == domain.MissionActive {
			cp := *x
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *memMissions) UpsertProgress(_ context.Context, p *domain.MissionProgress) error {
	cp := *p
	m.progress[m.key(p.MissionID, p.VolunteerID)] = &cp
	return nil
}
func (m *memMissions) GetProgress(_ context.Context, missionID, volunteerID uuid.UUID) (*domain.MissionProgress, error) {
	p, ok := m.progress[m.key(missionID, volunteerID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *p
	return &cp, nil
}
func (m *memMissions) ListProgressByVolunteer(_ context.Context, volunteerID uuid.UUID) ([]domain.MissionProgress, error) {
	var out []domain.MissionProgress
	for _, p := range m.progress {
		if p.VolunteerID == volunteerID {
			out = append(out, *p)
		}
	}
	return out, nil
}

type volDocs struct {
	memory.VolunteerAdapter
	docs   []domain.Document
	skills []domain.VolunteerSkill
}

func (v volDocs) ListDocuments(context.Context, uuid.UUID) ([]domain.Document, error) {
	return v.docs, nil
}
func (v volDocs) ListVolunteerSkills(context.Context, uuid.UUID) ([]domain.VolunteerSkill, error) {
	return v.skills, nil
}

func completeVolunteer(uid, vid uuid.UUID, status domain.VolunteerStatus) *domain.Volunteer {
	return &domain.Volunteer{
		ID: vid, UserID: uid, Status: status,
		FirstName: "سارا", LastName: "محمدی", FullName: "سارا محمدی",
		NationalID: "0012345678", Phone: "09121234567", BirthDate: "1996-05-12",
		Province: "تهران", City: "تهران", EducationLevel: "کارشناسی",
		SkillCategories: []domain.SkillCategory{domain.SkillArtistic},
	}
}

func TestVerifyRejectsIncompleteProfile(t *testing.T) {
	store := memory.New()
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), &domain.Volunteer{ID: vid, UserID: uid, Status: domain.StatusDraft, FirstName: "سارا"})
	mm := newMemMissions()
	svc := missionuc.New(mm, volDocs{VolunteerAdapter: memory.VolunteerAdapter{S: store}}, nil, domain.RealClock{}, nil)
	m, err := svc.Create(context.Background(), missionuc.MissionInput{
		Title: "پروفایل", Kind: domain.MissionCompleteProfile, HourWeight: 1, TargetCount: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Verify(context.Background(), uid, m.ID)
	if err == nil {
		t.Fatal("want not verified")
	}
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
	p, _ := mm.GetProgress(context.Background(), m.ID, vid)
	if p != nil && p.Status == domain.MissionCompleted {
		t.Fatal("must not complete")
	}
}

func TestVerifyCompletesSubmittedProfile(t *testing.T) {
	store := memory.New()
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), completeVolunteer(uid, vid, domain.StatusPending))
	mm := newMemMissions()
	vols := volDocs{
		VolunteerAdapter: memory.VolunteerAdapter{S: store},
		docs:             []domain.Document{{Kind: domain.DocNationalID}},
		skills:           []domain.VolunteerSkill{{SkillID: uuid.New()}},
	}
	svc := missionuc.New(mm, vols, nil, domain.RealClock{}, nil)
	m, err := svc.Create(context.Background(), missionuc.MissionInput{
		Title: "پروفایل", Kind: domain.MissionCompleteProfile, HourWeight: 1, TargetCount: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := svc.Verify(context.Background(), uid, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != domain.MissionCompleted {
		t.Fatalf("status=%s", p.Status)
	}
}

func TestInboundWebhookRequiredToAdvance(t *testing.T) {
	store := memory.New()
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), completeVolunteer(uid, vid, domain.StatusApproved))
	mm := newMemMissions()
	svc := missionuc.New(mm, memory.VolunteerAdapter{S: store}, nil, domain.RealClock{}, nil)
	m, err := svc.Create(context.Background(), missionuc.MissionInput{
		Title: "دعوت", Kind: domain.MissionInviteUsers, HourWeight: 2, TargetCount: 2, WebhookEvent: "user.invited",
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.VerifyToken == "" || m.VerifyMode != domain.VerifyInbound {
		t.Fatalf("token/mode %+v", m)
	}
	_, err = svc.Verify(context.Background(), uid, m.ID)
	if err == nil {
		t.Fatal("self-report must fail")
	}
	if err := svc.AwardInbound(context.Background(), "bad-token", "user.invited", vid.String(), "", 1); err != domain.ErrUnauthorized {
		t.Fatalf("want unauthorized, got %v", err)
	}
	if err := svc.AwardInbound(context.Background(), m.VerifyToken, "user.invited", "", "09121234567", 1); err != nil {
		t.Fatal(err)
	}
	p, _ := mm.GetProgress(context.Background(), m.ID, vid)
	if p.Progress != 1 || p.Status == domain.MissionCompleted {
		t.Fatalf("progress=%d status=%s", p.Progress, p.Status)
	}
	if err := svc.AwardInbound(context.Background(), m.VerifyToken, "user.invited", vid.String(), "", 1); err != nil {
		t.Fatal(err)
	}
	p, _ = mm.GetProgress(context.Background(), m.ID, vid)
	if p.Status != domain.MissionCompleted {
		t.Fatalf("status=%s progress=%d", p.Status, p.Progress)
	}
}

func TestOutboundHTTPVerify(t *testing.T) {
	store := memory.New()
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), completeVolunteer(uid, vid, domain.StatusApproved))
	mm := newMemMissions()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(raw), vid.String()) {
			t.Errorf("body=%s", raw)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "progress": 1})
	}))
	t.Cleanup(srv.Close)
	svc := missionuc.New(mm, memory.VolunteerAdapter{S: store}, nil, domain.RealClock{}, srv.Client())
	m, err := svc.Create(context.Background(), missionuc.MissionInput{
		Title: "سفارشی", Kind: domain.MissionCustom, HourWeight: 1, TargetCount: 1,
		VerifyMode: domain.VerifyOutbound, VerifyURL: srv.URL, VerifyToken: "secret-token",
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := svc.Verify(context.Background(), uid, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != domain.MissionCompleted {
		t.Fatalf("status=%s", p.Status)
	}
}

func TestOutboundHTTPNotYetDone(t *testing.T) {
	store := memory.New()
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), completeVolunteer(uid, vid, domain.StatusApproved))
	mm := newMemMissions()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "progress": 0, "message": "هنوز دعوت نشده"})
	}))
	t.Cleanup(srv.Close)
	svc := missionuc.New(mm, memory.VolunteerAdapter{S: store}, nil, domain.RealClock{}, srv.Client())
	m, err := svc.Create(context.Background(), missionuc.MissionInput{
		Title: "سفارشی", Kind: domain.MissionCustom, HourWeight: 1, TargetCount: 1,
		VerifyMode: domain.VerifyOutbound, VerifyURL: srv.URL, VerifyToken: "t",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Verify(context.Background(), uid, m.ID)
	if err == nil || !strings.Contains(err.Error(), "هنوز دعوت نشده") {
		t.Fatalf("got %v", err)
	}
}

func TestCreateOutboundRequiresURL(t *testing.T) {
	svc := missionuc.New(newMemMissions(), memory.VolunteerAdapter{S: memory.New()}, nil, domain.RealClock{}, nil)
	_, err := svc.Create(context.Background(), missionuc.MissionInput{
		Title: "سفارشی", Kind: domain.MissionCustom, HourWeight: 1, VerifyMode: domain.VerifyOutbound,
	})
	if err == nil {
		t.Fatal("want url error")
	}
}

func TestExpiredMission(t *testing.T) {
	store := memory.New()
	uid, vid := uuid.New(), uuid.New()
	_ = store.CreateVolunteer(context.Background(), completeVolunteer(uid, vid, domain.StatusApproved))
	mm := newMemMissions()
	clock := testClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	svc := missionuc.New(mm, memory.VolunteerAdapter{S: store}, nil, clock, nil)
	h := 1
	m, err := svc.Create(context.Background(), missionuc.MissionInput{
		Title: "دعوت", Kind: domain.MissionInviteUsers, HourWeight: 1, TargetCount: 1, DeadlineHours: &h,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Start(context.Background(), uid, m.ID); err != nil {
		t.Fatal(err)
	}
	svc2 := missionuc.New(mm, memory.VolunteerAdapter{S: store}, nil, testClock{t: clock.t.Add(2 * time.Hour)}, nil)
	_, err = svc2.Verify(context.Background(), uid, m.ID)
	if !errors.Is(err, domain.ErrMissionExpired) {
		t.Fatalf("got %v", err)
	}
}
