package volunteeruc

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

const maxDocumentBytes = 5 << 20 // 5MB

var allowedMime = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
	"application/pdf": {},
}

type Service struct {
	users      domain.UserRepository
	volunteers domain.VolunteerRepository
	storage    domain.ObjectStorage
	notify     domain.Notifier
	skills     domain.SkillRepository
	clock      domain.Clock
}

func New(users domain.UserRepository, volunteers domain.VolunteerRepository, storage domain.ObjectStorage, notify domain.Notifier, skills domain.SkillRepository, clock domain.Clock) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	return &Service{users: users, volunteers: volunteers, storage: storage, notify: notify, skills: skills, clock: clock}
}

type ProfileInput struct {
	FullName        string       `json:"full_name"`
	NationalID      string       `json:"national_id"`
	Phone           string       `json:"phone"`
	Phone2          string       `json:"phone2"`
	Province        string       `json:"province"`
	City            string       `json:"city"`
	Address         string       `json:"address"`
	Plaque          string       `json:"plaque"`
	Unit            string       `json:"unit"`
	Bio             string       `json:"bio"`
	SkillIDs        *[]uuid.UUID `json:"skill_ids"`
	SkillCategories []string     `json:"skill_categories"`
	EducationLevel  string       `json:"education_level"`
	EducationField  string       `json:"education_field"`
	MedicalLicense  string       `json:"medical_license"`
	BirthDate       string       `json:"birth_date"`
}

func (s *Service) UpsertProfile(ctx context.Context, userID uuid.UUID, in ProfileInput) (*domain.Volunteer, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	now := s.clock.Now()
	if err == domain.ErrNotFound {
		v = &domain.Volunteer{
			ID:              uuid.New(),
			UserID:          userID,
			Status:          domain.StatusDraft,
			CreatedAt:       now,
			SkillCategories: []domain.SkillCategory{},
		}
		applyProfile(v, in)
		v.UpdatedAt = now
		if err := s.volunteers.Create(ctx, v); err != nil {
			return nil, err
		}
		if err := s.applySkills(ctx, v, in.SkillIDs); err != nil {
			return nil, err
		}
		if err := s.volunteers.Update(ctx, v); err != nil {
			return nil, err
		}
		return s.hydrate(ctx, v)
	}
	if err != nil {
		return nil, err
	}
	applyProfile(v, in)
	if err := s.applySkills(ctx, v, in.SkillIDs); err != nil {
		return nil, err
	}
	if v.Status == domain.StatusRejected {
		v.Status = domain.StatusDraft
		v.RejectionReason = ""
	}
	v.UpdatedAt = now
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	return s.hydrate(ctx, v)
}

func applyProfile(v *domain.Volunteer, in ProfileInput) {
	if name := strings.TrimSpace(in.FullName); name != "" {
		v.FullName = name
	}
	v.NationalID = strings.TrimSpace(in.NationalID)
	v.Phone = strings.TrimSpace(in.Phone)
	v.Phone2 = strings.TrimSpace(in.Phone2)
	v.Province = strings.TrimSpace(in.Province)
	v.City = strings.TrimSpace(in.City)
	v.Address = strings.TrimSpace(in.Address)
	v.Plaque = strings.TrimSpace(in.Plaque)
	v.Unit = strings.TrimSpace(in.Unit)
	v.Bio = strings.TrimSpace(in.Bio)
	if in.SkillIDs == nil {
		v.SkillCategories = domain.ParseSkillCategories(in.SkillCategories)
	}
	v.EducationLevel = strings.TrimSpace(in.EducationLevel)
	v.EducationField = strings.TrimSpace(in.EducationField)
	v.MedicalLicense = strings.TrimSpace(in.MedicalLicense)
	v.BirthDate = strings.TrimSpace(in.BirthDate)
}

func (s *Service) applySkills(ctx context.Context, v *domain.Volunteer, ids *[]uuid.UUID) error {
	if ids == nil {
		return nil
	}
	if err := s.volunteers.ReplaceSkills(ctx, v.ID, *ids); err != nil {
		return err
	}
	return s.syncCategories(ctx, v)
}

func (s *Service) syncCategories(ctx context.Context, v *domain.Volunteer) error {
	list, err := s.volunteers.ListVolunteerSkills(ctx, v.ID)
	if err != nil {
		return err
	}
	seen := map[string]struct{}{}
	var cats []string
	for _, sk := range list {
		slug := strings.TrimSpace(sk.GroupSlug)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		cats = append(cats, slug)
	}
	v.SkillCategories = domain.ParseSkillCategories(cats)
	v.Skills = list
	return nil
}

func (s *Service) hydrate(ctx context.Context, v *domain.Volunteer) (*domain.Volunteer, error) {
	if v == nil {
		return nil, domain.ErrNotFound
	}
	if skills, err := s.volunteers.ListVolunteerSkills(ctx, v.ID); err == nil {
		v.Skills = skills
	}
	if v.Skills == nil {
		v.Skills = []domain.VolunteerSkill{}
	}
	if s.skills != nil {
		if props, err := s.skills.ListProposalsByVolunteer(ctx, v.ID); err == nil {
			v.Proposals = props
		}
	}
	if v.Proposals == nil {
		v.Proposals = []domain.SkillProposal{}
	}
	return v, nil
}

func (s *Service) GetMine(ctx context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.hydrate(ctx, v)
}

func (s *Service) SubmitForReview(ctx context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	switch {
	case strings.TrimSpace(v.FullName) == "":
		return nil, domain.Invalid("نام کامل الزامی است")
	case strings.TrimSpace(v.NationalID) == "":
		return nil, domain.Invalid("کد ملی الزامی است")
	case strings.TrimSpace(v.Phone) == "":
		return nil, domain.Invalid("شماره موبایل الزامی است")
	case strings.TrimSpace(v.BirthDate) == "":
		return nil, domain.Invalid("تاریخ تولد را وارد کنید")
	case strings.TrimSpace(v.Province) == "":
		return nil, domain.Invalid("استان را انتخاب کنید")
	case strings.TrimSpace(v.City) == "":
		return nil, domain.Invalid("شهر را انتخاب کنید")
	case strings.TrimSpace(v.EducationLevel) == "":
		return nil, domain.Invalid("میزان تحصیلات را انتخاب کنید")
	}
	docs, err := s.volunteers.ListDocuments(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	hasNational := false
	for _, d := range docs {
		if d.Kind == domain.DocNationalID {
			hasNational = true
			break
		}
	}
	if !hasNational {
		return nil, domain.ErrDocumentRequired
	}
	skills, err := s.volunteers.ListVolunteerSkills(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	hasPendingProposal := false
	if s.skills != nil {
		props, err := s.skills.ListProposalsByVolunteer(ctx, v.ID)
		if err != nil {
			return nil, err
		}
		for _, p := range props {
			if p.Status == domain.ProposalPending || p.Status == domain.ProposalApproved {
				hasPendingProposal = true
				break
			}
		}
	}
	if len(skills) == 0 && !hasPendingProposal && len(v.SkillCategories) == 0 {
		return nil, domain.Invalid("حداقل یک مهارت انتخاب کنید یا مهارت جدید پیشنهاد دهید")
	}
	if !domain.CanTransition(v.Status, domain.StatusPending) {
		return nil, domain.ErrInvalidTransition
	}
	v.Status = domain.StatusPending
	v.RejectionReason = ""
	v.UpdatedAt = s.clock.Now()
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	return s.hydrate(ctx, v)
}

func (s *Service) SetAvailability(ctx context.Context, userID uuid.UUID, slots []domain.AvailabilitySlot) error {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for i := range slots {
		if slots[i].Weekday < 0 || slots[i].Weekday > 6 {
			return domain.ErrInvalidInput
		}
		slots[i].ID = uuid.New()
		slots[i].VolunteerID = v.ID
	}
	return s.volunteers.ReplaceAvailability(ctx, v.ID, slots)
}

func (s *Service) ListAvailability(ctx context.Context, volunteerID uuid.UUID) ([]domain.AvailabilitySlot, error) {
	return s.volunteers.ListAvailability(ctx, volunteerID)
}

func (s *Service) UploadDocument(ctx context.Context, userID uuid.UUID, kind domain.DocumentKind, fileName, mime string, size int64, body io.Reader) (*domain.Document, error) {
	if size <= 0 || size > maxDocumentBytes {
		return nil, domain.ErrFileTooLarge
	}
	if _, ok := allowedMime[mime]; !ok {
		return nil, domain.ErrInvalidFileType
	}
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	ext := path.Ext(fileName)
	key := fmt.Sprintf("volunteers/%s/%s-%s%s", v.ID, kind, uuid.NewString(), ext)
	if err := s.storage.Put(ctx, key, body, size, mime); err != nil {
		return nil, err
	}
	doc := &domain.Document{
		ID:          uuid.New(),
		VolunteerID: v.ID,
		Kind:        kind,
		ObjectKey:   key,
		FileName:    fileName,
		MimeType:    mime,
		SizeBytes:   size,
		CreatedAt:   s.clock.Now(),
	}
	if err := s.volunteers.AddDocument(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *Service) ListDocuments(ctx context.Context, volunteerID uuid.UUID) ([]domain.Document, error) {
	return s.volunteers.ListDocuments(ctx, volunteerID)
}

func (s *Service) Review(ctx context.Context, actorID, volunteerID uuid.UUID, action, reason string) (*domain.Volunteer, error) {
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	var next domain.VolunteerStatus
	title, body := "", ""
	switch action {
	case "approve":
		next = domain.StatusApproved
		title = "تایید عضویت داوطلبی"
		body = "پروفایل شما تایید شد. از این پس می‌توانید تسک‌های عملیاتی را مشاهده و پذیرش کنید."
	case "reject":
		if strings.TrimSpace(reason) == "" {
			return nil, domain.ErrInvalidInput
		}
		next = domain.StatusRejected
		v.RejectionReason = reason
		title = "رد درخواست داوطلبی"
		body = "درخواست شما رد شد. دلیل: " + reason
	case "request_documents":
		if strings.TrimSpace(reason) == "" {
			return nil, domain.ErrInvalidInput
		}
		next = domain.StatusDraft
		v.RejectionReason = reason
		title = "نیاز به تکمیل مدارک"
		body = "لطفا مدارک را اصلاح کنید: " + reason
	case "suspend":
		next = domain.StatusSuspended
		v.RejectionReason = reason
		title = "تعلیق حساب داوطلبی"
		body = "حساب شما موقتا تعلیق شد. " + reason
	case "unsuspend":
		next = domain.StatusApproved
		v.RejectionReason = ""
		title = "رفع تعلیق"
		body = "تعلیق حساب شما برداشته شد. می‌توانید دوباره در فعالیت‌ها شرکت کنید."
	default:
		return nil, domain.ErrInvalidInput
	}
	if !domain.CanTransition(v.Status, next) {
		return nil, domain.ErrInvalidTransition
	}
	v.Status = next
	v.UpdatedAt = s.clock.Now()
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	if s.notify != nil {
		_ = s.notify.Notify(ctx, v.UserID, title, body)
	}
	_ = actorID
	return v, nil
}

func (s *Service) List(ctx context.Context, f domain.VolunteerFilter) ([]domain.Volunteer, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.volunteers.List(ctx, f)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Volunteer, error) {
	v, err := s.volunteers.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.hydrate(ctx, v)
}

func (s *Service) OpenDocument(ctx context.Context, id uuid.UUID) (io.ReadCloser, *domain.Document, error) {
	doc, err := s.volunteers.GetDocument(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	r, _, err := s.storage.Get(ctx, doc.ObjectKey)
	if err != nil {
		return nil, nil, err
	}
	return r, doc, nil
}

func WeekdayName(d int) string {
	names := []string{"یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه", "شنبه"}
	if d < 0 || d > 6 {
		return ""
	}
	return names[d]
}

func FormatClock(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
