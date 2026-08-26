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

type staffNotifier interface {
	NotifyStaff(ctx context.Context, title, body string) error
}

func (s *Service) notifyStaff(ctx context.Context, title, body string) {
	if sn, ok := s.notify.(staffNotifier); ok {
		_ = sn.NotifyStaff(ctx, title, body)
	}
}

type ProfileInput struct {
	FullName        string       `json:"full_name"`
	FirstName       string       `json:"first_name"`
	LastName        string       `json:"last_name"`
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
	Gender          string       `json:"gender"`
	Occupation      string       `json:"occupation"`
	OccupationOther string       `json:"occupation_other"`
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
		if s.users != nil {
			if u, uerr := s.users.GetByID(ctx, userID); uerr == nil {
				v.Phone = u.Phone
			}
		}
		if err := applyProfile(v, in, false, now); err != nil {
			return nil, err
		}
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
	locked := identityLocked(v.Status)
	if err := applyProfile(v, in, locked, now); err != nil {
		return nil, err
	}
	if s.users != nil {
		if u, uerr := s.users.GetByID(ctx, userID); uerr == nil && u.Phone != "" {
			v.Phone = u.Phone
		}
	}
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

func identityLocked(status domain.VolunteerStatus) bool {
	return status == domain.StatusApproved || status == domain.StatusPending || status == domain.StatusSuspended
}

func applyProfile(v *domain.Volunteer, in ProfileInput, identityLocked bool, now time.Time) error {
	if !identityLocked {
		first, last := splitName(in.FirstName, in.LastName, in.FullName)
		if first != "" {
			if err := validatePersianName(first); err != nil {
				return err
			}
			v.FirstName = first
		}
		if last != "" {
			if err := validatePersianName(last); err != nil {
				return err
			}
			v.LastName = last
		}
		if v.FirstName != "" || v.LastName != "" {
			v.FullName = strings.TrimSpace(v.FirstName + " " + v.LastName)
		}
		nid := normalizeDigits(in.NationalID)
		if nid != "" {
			if err := validateNationalID(nid); err != nil {
				return err
			}
			v.NationalID = nid
		}
		if bd := strings.TrimSpace(in.BirthDate); bd != "" {
			if err := validateBirthDate(bd, now); err != nil {
				return err
			}
			v.BirthDate = bd
		}
		if err := applyGenderOccupation(v, in); err != nil {
			return err
		}
	}
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
	return nil
}

func (s *Service) AdminUpdate(ctx context.Context, actorID, volunteerID uuid.UUID, in ProfileInput) (*domain.Volunteer, error) {
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := applyProfile(v, in, false, now); err != nil {
		return nil, err
	}
	if phone := strings.TrimSpace(in.Phone); phone != "" {
		v.Phone = phone
	}
	if err := s.applySkills(ctx, v, in.SkillIDs); err != nil {
		return nil, err
	}
	v.UpdatedAt = now
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	s.addEvent(ctx, v.ID, actorID, "admin", domain.EventProfileUpdated, v.Status, v.Status, "اطلاعات پرونده توسط ادمین ویرایش شد")
	return s.hydrate(ctx, v)
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
	if v.Email == "" && s.users != nil {
		if u, uerr := s.users.GetByID(ctx, v.UserID); uerr == nil {
			v.Email = u.Email
		}
	}
	if events, err := s.volunteers.ListEvents(ctx, v.ID, 100); err == nil {
		v.History = events
	}
	if v.History == nil {
		v.History = []domain.VolunteerEvent{}
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
	var birthErr error
	if strings.TrimSpace(v.BirthDate) != "" {
		birthErr = validateBirthDate(v.BirthDate, s.clock.Now())
	}
	switch {
	case strings.TrimSpace(v.FirstName) == "" && strings.TrimSpace(v.FullName) == "":
		return nil, domain.Invalid("نام را وارد کنید")
	case strings.TrimSpace(v.LastName) == "" && !strings.Contains(strings.TrimSpace(v.FullName), " "):
		return nil, domain.Invalid("نام خانوادگی را وارد کنید")
	case strings.TrimSpace(v.NationalID) == "":
		return nil, domain.Invalid("کد ملی الزامی است")
	case validateNationalID(normalizeDigits(v.NationalID)) != nil:
		return nil, domain.Invalid("کد ملی باید ۱۰ رقم باشد")
	case strings.TrimSpace(v.Phone) == "":
		return nil, domain.Invalid("شماره موبایل الزامی است")
	case strings.TrimSpace(v.BirthDate) == "":
		return nil, domain.Invalid("تاریخ تولد را وارد کنید")
	case birthErr != nil:
		return nil, birthErr
	case !validGender(v.Gender):
		return nil, domain.Invalid("جنسیت را انتخاب کنید")
	case !validOccupation(v.Occupation):
		return nil, domain.Invalid("شغل را انتخاب کنید")
	case v.Occupation == occupationOther && strings.TrimSpace(v.OccupationOther) == "":
		return nil, domain.Invalid("در صورت انتخاب «سایر»، شغل خود را بنویسید")
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
	from := v.Status
	v.Status = domain.StatusPending
	v.RejectionReason = ""
	v.UpdatedAt = s.clock.Now()
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	s.addEvent(ctx, v.ID, userID, "volunteer", domain.EventSubmitted, from, domain.StatusPending, "درخواست برای بررسی ادمین ارسال شد")
	title := "ارسال پرونده برای بررسی"
	if from == domain.StatusDraft || from == domain.StatusRejected {
		title = "ارسال مجدد پرونده پس از نقص مدرک"
	}
	s.notifyStaff(ctx, title, v.FullName+" پرونده را برای بررسی ادمین ارسال کرد.")
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
	if v.Status == domain.StatusDraft || v.Status == domain.StatusRejected {
		s.addEvent(ctx, v.ID, userID, "volunteer", domain.EventDocumentUploaded, v.Status, v.Status, "مدرک دوباره بارگذاری شد")
		s.notifyStaff(ctx, "بارگذاری مجدد مدارک", v.FullName+" مدارک را پس از نقص مدرک دوباره بارگذاری کرد.")
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
	reason = strings.TrimSpace(reason)
	var next domain.VolunteerStatus
	title, body := "", ""
	eventType := domain.EventStatusChanged
	switch action {
	case "approve":
		next = domain.StatusApproved
		eventType = domain.EventApproved
		title = "تایید عضویت داوطلبی"
		body = "پروفایل شما تایید شد. از این پس می‌توانید فعالیت‌های عملیاتی را مشاهده و درخواست دهید."
	case "reject":
		if reason == "" {
			return nil, domain.Invalid("برای رد کردن درخواست باید دلیل ثبت شود")
		}
		next = domain.StatusRejected
		eventType = domain.EventRejected
		v.RejectionReason = reason
		title = "رد درخواست داوطلبی"
		body = "درخواست شما رد شد. دلیل: " + reason
	case "request_documents":
		if reason == "" {
			return nil, domain.Invalid("توضیح مدارک درخواستی الزامی است")
		}
		next = domain.StatusDraft
		eventType = domain.EventDocumentsRequested
		v.RejectionReason = reason
		title = "نیاز به تکمیل مدارک"
		body = "لطفا مدارک را اصلاح کنید: " + reason
	case "suspend":
		next = domain.StatusSuspended
		eventType = domain.EventSuspended
		v.RejectionReason = reason
		title = "تعلیق حساب داوطلبی"
		body = "حساب شما موقتا تعلیق شد. " + reason
	case "unsuspend":
		next = domain.StatusApproved
		eventType = domain.EventUnsuspended
		v.RejectionReason = ""
		title = "رفع تعلیق"
		body = "تعلیق حساب شما برداشته شد. می‌توانید دوباره در فعالیت‌ها شرکت کنید."
	default:
		return nil, domain.Invalid("عملیات نامعتبر است")
	}
	if !domain.CanTransition(v.Status, next) {
		return nil, domain.ErrInvalidTransition
	}
	from := v.Status
	v.Status = next
	v.UpdatedAt = s.clock.Now()
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	s.addEvent(ctx, v.ID, actorID, "admin", eventType, from, next, reason)
	if s.notify != nil && action != "suspend" && action != "unsuspend" {
		_ = s.notify.Notify(ctx, v.UserID, title, body)
	}
	return s.hydrate(ctx, v)
}

func (s *Service) SetStatus(ctx context.Context, actorID, volunteerID uuid.UUID, status, reason string) (*domain.Volunteer, error) {
	next, ok := domain.ParseVolunteerStatus(status)
	if !ok {
		return nil, domain.Invalid("وضعیت نامعتبر است")
	}
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	reason = strings.TrimSpace(reason)
	from := v.Status
	if reason == "" {
		if next == domain.StatusRejected {
			return nil, domain.Invalid("برای رد کردن درخواست باید دلیل ثبت شود")
		}
		return nil, domain.Invalid("برای تغییر وضعیت باید دلیل ثبت شود")
	}
	if from == next {
		if reason == "" {
			return s.hydrate(ctx, v)
		}
		s.addEvent(ctx, v.ID, actorID, "admin", domain.EventComment, from, next, reason)
		if s.notify != nil {
			_ = s.notify.Notify(ctx, v.UserID, "پیام پرونده داوطلبی", reason)
		}
		return s.hydrate(ctx, v)
	}
	switch next {
	case domain.StatusRejected:
		v.RejectionReason = reason
	case domain.StatusApproved:
		v.RejectionReason = ""
	default:
		if reason != "" {
			v.RejectionReason = reason
		}
	}
	v.Status = next
	v.UpdatedAt = s.clock.Now()
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	eventType := domain.EventStatusChanged
	switch next {
	case domain.StatusRejected:
		eventType = domain.EventRejected
	case domain.StatusApproved:
		eventType = domain.EventApproved
	case domain.StatusDraft:
		if reason != "" {
			eventType = domain.EventDocumentsRequested
		}
	case domain.StatusSuspended:
		eventType = domain.EventSuspended
	}
	s.addEvent(ctx, v.ID, actorID, "admin", eventType, from, next, reason)
	if s.notify != nil && next != domain.StatusSuspended && from != domain.StatusSuspended {
		title, body := statusNotify(next, reason)
		_ = s.notify.Notify(ctx, v.UserID, title, body)
	}
	return s.hydrate(ctx, v)
}

func (s *Service) AddComment(ctx context.Context, actorID, volunteerID uuid.UUID, comment string) (*domain.Volunteer, error) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return nil, domain.Invalid("متن پیام را وارد کنید")
	}
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	s.addEvent(ctx, v.ID, actorID, "admin", domain.EventComment, v.Status, v.Status, comment)
	if s.notify != nil {
		_ = s.notify.Notify(ctx, v.UserID, "پیام ادمین", comment)
	}
	return s.hydrate(ctx, v)
}

func (s *Service) DeleteMyDocument(ctx context.Context, userID, documentID uuid.UUID) error {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if v.Status == domain.StatusApproved || v.Status == domain.StatusSuspended {
		return domain.Invalid("پس از بررسی وضعیت توسط ادمین امکان حذف مدرک وجود ندارد")
	}
	doc, err := s.volunteers.GetDocument(ctx, documentID)
	if err != nil {
		return err
	}
	if doc.VolunteerID != v.ID {
		return domain.ErrForbidden
	}
	if err := s.volunteers.DeleteDocument(ctx, documentID); err != nil {
		return err
	}
	if s.storage != nil && doc.ObjectKey != "" {
		_ = s.storage.Delete(ctx, doc.ObjectKey)
	}
	s.addEvent(ctx, v.ID, userID, "volunteer", domain.EventDocumentDeleted, v.Status, v.Status, string(doc.Kind)+" — "+doc.FileName)
	return nil
}

func (s *Service) addEvent(ctx context.Context, volunteerID, actorID uuid.UUID, role string, typ domain.VolunteerEventType, from, to domain.VolunteerStatus, comment string) {
	_ = s.volunteers.AddEvent(ctx, &domain.VolunteerEvent{
		ID:          uuid.New(),
		VolunteerID: volunteerID,
		ActorUserID: actorID,
		ActorRole:   role,
		EventType:   typ,
		FromStatus:  from,
		ToStatus:    to,
		Comment:     strings.TrimSpace(comment),
		CreatedAt:   s.clock.Now(),
	})
}

func statusNotify(status domain.VolunteerStatus, reason string) (string, string) {
	switch status {
	case domain.StatusApproved:
		return "تایید عضویت داوطلبی", "پروفایل شما تایید شد. از این پس می‌توانید فعالیت‌های عملیاتی را مشاهده و درخواست دهید."
	case domain.StatusRejected:
		return "رد درخواست داوطلبی", "درخواست شما رد شد. دلیل: " + reason
	case domain.StatusDraft:
		if reason != "" {
			return "نیاز به تکمیل مدارک", "لطفا مدارک را اصلاح کنید: " + reason
		}
		return "بازگشت به پیش‌نویس", "پرونده شما به پیش‌نویس برگشت تا بتوانید آن را تکمیل کنید."
	case domain.StatusPending:
		return "وضعیت پرونده", "پرونده شما در انتظار بررسی ادمین قرار گرفت."
	case domain.StatusSuspended:
		return "تعلیق حساب داوطلبی", "حساب شما موقتا تعلیق شد. " + reason
	default:
		return "تغییر وضعیت پرونده", reason
	}
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
