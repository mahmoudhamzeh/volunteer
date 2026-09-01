package certuc

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type Service struct {
	certs      domain.CertificateRepository
	tasks      domain.TaskRepository
	volunteers domain.VolunteerRepository
	notify     domain.Notifier
	clock      domain.Clock
	publicBase string
}

func New(certs domain.CertificateRepository, tasks domain.TaskRepository, volunteers domain.VolunteerRepository, notify domain.Notifier, clock domain.Clock, publicBase string) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	if publicBase == "" {
		publicBase = "http://localhost:3000"
	}
	return &Service{certs: certs, tasks: tasks, volunteers: volunteers, notify: notify, clock: clock, publicBase: publicBase}
}

func (s *Service) IssueForAssignment(ctx context.Context, assignmentID uuid.UUID) (*domain.Certificate, error) {
	c, created, err := s.issueForAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if created {
		s.notifyVolunteer(ctx, c.VolunteerID, "تقدیرنامه صادر شد", "تقدیرنامه فعالیت «"+c.Title+"» صادر شد. از صفحه تقدیرنامه و گواهی می‌توانید PDF را دانلود کنید.")
	}
	return c, nil
}

func (s *Service) issueForAssignment(ctx context.Context, assignmentID uuid.UUID) (*domain.Certificate, bool, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, false, err
	}
	if !a.CanIssueCertificate() {
		return nil, false, domain.Invalid("این فعالیت هنوز تکمیل نشده است؛ پس از ثبت امتیاز می‌توان تقدیرنامه صادر کرد")
	}
	if existing, err := s.certs.GetByAssignment(ctx, assignmentID); err == nil && existing != nil {
		if v, verr := s.volunteers.GetByID(ctx, existing.VolunteerID); verr == nil {
			existing.Volunteer = v
		}
		return existing, false, nil
	} else if err != nil && err != domain.ErrNotFound {
		return nil, false, err
	}
	v, err := s.volunteers.GetByID(ctx, a.VolunteerID)
	if err != nil {
		return nil, false, err
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, false, err
	}
	hours := a.HoursAwarded
	if hours <= 0 {
		hours = t.HourWeight
	}
	now := s.clock.Now()
	c := &domain.Certificate{
		ID:               uuid.New(),
		VerificationCode: uuid.New(),
		VolunteerID:      v.ID,
		Kind:             domain.CertTask,
		AssignmentID:     &a.ID,
		Title:            t.Title,
		Hours:            hours,
		IssuedAt:         now,
		Volunteer:        v,
	}
	if err := s.certs.Create(ctx, c); err != nil {
		return nil, false, err
	}
	return c, true, nil
}

func (s *Service) IssueAggregated(ctx context.Context, volunteerID uuid.UUID, from, to time.Time) (*domain.Certificate, error) {
	c, err := s.issueAggregated(ctx, volunteerID, from, to)
	if err != nil {
		return nil, err
	}
	s.notifyVolunteer(ctx, volunteerID, "تقدیرنامه تجمیعی صادر شد", "تقدیرنامه تجمیعی همکاری داوطلبانه صادر شد. از صفحه تقدیرنامه و گواهی می‌توانید PDF را دانلود کنید.")
	return c, nil
}

func (s *Service) issueAggregated(ctx context.Context, volunteerID uuid.UUID, from, to time.Time) (*domain.Certificate, error) {
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	if v.CompletedTasks <= 0 && v.TotalHours <= 0 {
		return nil, domain.Invalid("برای صدور تقدیرنامه تجمیعی باید حداقل یک فعالیت تکمیل‌شده وجود داشته باشد")
	}
	now := s.clock.Now()
	title := "تقدیرنامه همکاری داوطلبانه"
	c := &domain.Certificate{
		ID:               uuid.New(),
		VerificationCode: uuid.New(),
		VolunteerID:      v.ID,
		Kind:             domain.CertAggregated,
		Title:            title,
		Hours:            v.TotalHours,
		PeriodStart:      &from,
		PeriodEnd:        &to,
		IssuedAt:         now,
		Volunteer:        v,
	}
	return c, s.certs.Create(ctx, c)
}

func (s *Service) Request(ctx context.Context, userID uuid.UUID, kind domain.CertificateKind, assignmentID *uuid.UUID) (*domain.CertificateRequest, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if v.Status != domain.StatusApproved {
		return nil, domain.ErrNotApproved
	}
	if kind != domain.CertTask && kind != domain.CertAggregated && kind != domain.CertOfficial {
		return nil, domain.Invalid("نوع مدرک نامعتبر است")
	}
	req := &domain.CertificateRequest{
		ID:            uuid.New(),
		VolunteerID:   v.ID,
		VolunteerName: v.FullName,
		Kind:          kind,
		Status:        domain.CertReqPending,
		CreatedAt:     s.clock.Now(),
	}
	if kind == domain.CertOfficial {
		if v.TotalHours+0.0001 < domain.MinOfficialHours {
			return nil, domain.Invalid(fmt.Sprintf("برای درخواست گواهی‌نامه فعالیت داوطلبانه باید حداقل %d ساعت فعالیت تاییدشده داشته باشید", domain.MinOfficialHours))
		}
		pending, err := s.certs.HasPendingRequest(ctx, v.ID, domain.CertOfficial, nil)
		if err != nil {
			return nil, err
		}
		if pending {
			return nil, domain.Invalid("درخواست گواهی‌نامه شما در حال آماده‌سازی یا تحویل است")
		}
		req.Status = domain.CertReqPreparing
	} else if kind == domain.CertTask {
		if assignmentID == nil || *assignmentID == uuid.Nil {
			return nil, domain.Invalid("فعالیت تکمیل‌شده را برای صدور تقدیرنامه انتخاب کنید")
		}
		a, err := s.tasks.GetAssignment(ctx, *assignmentID)
		if err != nil {
			return nil, err
		}
		if a.VolunteerID != v.ID {
			return nil, domain.ErrForbidden
		}
		if !a.CanIssueCertificate() {
			return nil, domain.Invalid("فقط برای فعالیت تکمیل‌شده می‌توان تقدیرنامه درخواست کرد")
		}
		if existing, err := s.certs.GetByAssignment(ctx, a.ID); err == nil && existing != nil {
			return nil, domain.Invalid("تقدیرنامه این فعالیت قبلاً صادر شده است")
		} else if err != nil && err != domain.ErrNotFound {
			return nil, err
		}
		pending, err := s.certs.HasPendingRequest(ctx, v.ID, domain.CertTask, &a.ID)
		if err != nil {
			return nil, err
		}
		if pending {
			return nil, domain.Invalid("درخواست تقدیرنامه این فعالیت قبلاً ثبت شده و در حال بررسی است")
		}
		req.AssignmentID = &a.ID
		if t, terr := s.tasks.GetByID(ctx, a.TaskID); terr == nil {
			req.AssignmentTitle = t.Title
		}
	} else {
		if v.CompletedTasks <= 0 && v.TotalHours <= 0 {
			return nil, domain.Invalid("برای درخواست تقدیرنامه تجمیعی باید حداقل یک فعالیت تکمیل‌شده داشته باشید")
		}
		pending, err := s.certs.HasPendingRequest(ctx, v.ID, domain.CertAggregated, nil)
		if err != nil {
			return nil, err
		}
		if pending {
			return nil, domain.Invalid("درخواست تقدیرنامه تجمیعی قبلاً ثبت شده و در حال بررسی است")
		}
	}
	if err := s.certs.CreateRequest(ctx, req); err != nil {
		return nil, err
	}
	s.recordEvent(ctx, v, "volunteer", userID, "درخواست «"+requestTitle(*req)+"» ثبت شد")
	if sn, ok := s.notify.(interface {
		NotifyStaff(ctx context.Context, title, body string) error
	}); ok {
		_ = sn.NotifyStaff(ctx, "درخواست مدرک همکاری", v.FullName+" درخواست «"+requestTitle(*req)+"» ثبت کرد.")
	}
	if kind == domain.CertOfficial {
		s.notifyVolunteer(ctx, v.ID, "درخواست گواهی‌نامه ثبت شد", "درخواست گواهی‌نامه فعالیت داوطلبانه شما ثبت شد و در حال آماده‌سازی است.")
	}
	return req, nil
}

func (s *Service) ListMyRequests(ctx context.Context, userID uuid.UUID) ([]domain.CertificateRequest, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.certs.ListRequestsByVolunteer(ctx, v.ID)
}

func (s *Service) ListRequests(ctx context.Context, status domain.CertificateRequestStatus) ([]domain.CertificateRequest, error) {
	return s.certs.ListRequests(ctx, status)
}

func (s *Service) ReviewRequest(ctx context.Context, requestID uuid.UUID, action, note string) (*domain.CertificateRequest, error) {
	return s.Review(ctx, requestID, ReviewInput{Action: action, Note: note})
}

type ReviewInput struct {
	Action         string
	Note           string
	DeliveryMethod string
}

func (s *Service) Review(ctx context.Context, requestID uuid.UUID, in ReviewInput) (*domain.CertificateRequest, error) {
	req, err := s.certs.GetRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	note := strings.TrimSpace(in.Note)
	now := s.clock.Now()
	action := strings.TrimSpace(strings.ToLower(in.Action))
	if req.Kind == domain.CertOfficial {
		return s.reviewOfficial(ctx, req, action, note, strings.TrimSpace(strings.ToLower(in.DeliveryMethod)), now)
	}
	if req.Status != domain.CertReqPending {
		return nil, domain.Invalid("این درخواست قبلاً بررسی شده است")
	}
	req.AdminNote = note
	req.ReviewedAt = &now
	switch action {
	case "approve":
		var c *domain.Certificate
		if req.Kind == domain.CertTask {
			if req.AssignmentID == nil {
				return nil, domain.Invalid("فعالیت این درخواست نامعتبر است")
			}
			issued, created, err := s.issueForAssignment(ctx, *req.AssignmentID)
			if err != nil {
				return nil, err
			}
			c = issued
			if created {
				s.notifyVolunteer(ctx, req.VolunteerID, "تقدیرنامه صادر شد", "درخواست تقدیرنامه شما تایید شد. تقدیرنامه «"+c.Title+"» صادر شد و در صفحه تقدیرنامه و گواهی در دسترس است.")
			} else {
				s.notifyVolunteer(ctx, req.VolunteerID, "تقدیرنامه صادر شد", "تقدیرنامه «"+c.Title+"» آماده است. از صفحه تقدیرنامه و گواهی PDF را دانلود کنید.")
			}
		} else {
			c, err = s.issueAggregated(ctx, req.VolunteerID, now.AddDate(-1, 0, 0), now)
			if err != nil {
				return nil, err
			}
			s.notifyVolunteer(ctx, req.VolunteerID, "تقدیرنامه تجمیعی صادر شد", "درخواست تقدیرنامه تجمیعی شما تایید شد. از صفحه تقدیرنامه و گواهی می‌توانید PDF را دانلود کنید.")
		}
		req.Status = domain.CertReqApproved
		req.CertificateID = &c.ID
	case "reject":
		if note == "" {
			return nil, domain.Invalid("برای رد درخواست تقدیرنامه دلیل را بنویسید")
		}
		req.Status = domain.CertReqRejected
		s.notifyVolunteer(ctx, req.VolunteerID, "درخواست تقدیرنامه رد شد", "درخواست تقدیرنامه رد شد. دلیل: "+note)
	default:
		return nil, domain.Invalid("عملیات نامعتبر است؛ تایید یا رد را انتخاب کنید")
	}
	if err := s.certs.UpdateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) reviewOfficial(ctx context.Context, req *domain.CertificateRequest, action, note, deliveryMethod string, now time.Time) (*domain.CertificateRequest, error) {
	switch action {
	case "approve":
		if req.Status != domain.CertReqPreparing {
			return nil, domain.Invalid("فقط درخواست در حال آماده‌سازی را می‌توان صادر کرد")
		}
		c, err := s.issueOfficial(ctx, req.VolunteerID)
		if err != nil {
			return nil, err
		}
		req.Status = domain.CertReqReady
		req.CertificateID = &c.ID
		req.AdminNote = note
		req.ReviewedAt = &now
		s.notifyVolunteer(ctx, req.VolunteerID, "گواهی‌نامه آماده تحویل است", "گواهی‌نامه فعالیت داوطلبانه شما صادر شد و آماده تحویل است. ارسال یا دریافت حضوری را واحد پشتیبانی هماهنگ می‌کند.")
	case "reject":
		if req.Status != domain.CertReqPreparing {
			return nil, domain.Invalid("فقط درخواست در حال آماده‌سازی را می‌توان رد کرد")
		}
		if note == "" {
			return nil, domain.Invalid("برای رد درخواست گواهی‌نامه دلیل را بنویسید")
		}
		req.Status = domain.CertReqRejected
		req.AdminNote = note
		req.ReviewedAt = &now
		s.notifyVolunteer(ctx, req.VolunteerID, "درخواست گواهی‌نامه رد شد", "درخواست گواهی‌نامه فعالیت داوطلبانه رد شد. دلیل: "+note)
	case "deliver":
		if req.Status != domain.CertReqReady {
			return nil, domain.Invalid("ابتدا گواهی‌نامه باید صادر و آماده تحویل شود")
		}
		if deliveryMethod != "send" && deliveryMethod != "in_person" {
			return nil, domain.Invalid("نحوه تحویل را مشخص کنید: ارسال یا حضوری")
		}
		req.Status = domain.CertReqDelivered
		req.DeliveryMethod = deliveryMethod
		req.DeliveredAt = &now
		if note != "" {
			req.AdminNote = note
		}
		how := "ارسال شد"
		if deliveryMethod == "in_person" {
			how = "حضوری تحویل شد"
		}
		s.notifyVolunteer(ctx, req.VolunteerID, "گواهی‌نامه تحویل شد", "گواهی‌نامه فعالیت داوطلبانه شما "+how+".")
	default:
		return nil, domain.Invalid("عملیات نامعتبر است")
	}
	if err := s.certs.UpdateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) issueOfficial(ctx context.Context, volunteerID uuid.UUID) (*domain.Certificate, error) {
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	if v.TotalHours+0.0001 < domain.MinOfficialHours {
		return nil, domain.Invalid(fmt.Sprintf("برای صدور گواهی‌نامه باید حداقل %d ساعت فعالیت تاییدشده ثبت شده باشد", domain.MinOfficialHours))
	}
	now := s.clock.Now()
	c := &domain.Certificate{
		ID:               uuid.New(),
		VerificationCode: uuid.New(),
		VolunteerID:      v.ID,
		Kind:             domain.CertOfficial,
		Title:            "گواهی‌نامه فعالیت داوطلبانه",
		Hours:            v.TotalHours,
		IssuedAt:         now,
		Volunteer:        v,
	}
	return c, s.certs.Create(ctx, c)
}

func requestTitle(req domain.CertificateRequest) string {
	if req.Kind == domain.CertOfficial {
		return "گواهی‌نامه فعالیت داوطلبانه"
	}
	if req.AssignmentTitle != "" {
		return req.AssignmentTitle
	}
	if req.Kind == domain.CertAggregated {
		return "تقدیرنامه تجمیعی"
	}
	return "تقدیرنامه فعالیت"
}

func (s *Service) notifyVolunteer(ctx context.Context, volunteerID uuid.UUID, title, body string) {
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return
	}
	if s.notify != nil {
		_ = s.notify.Notify(ctx, v.UserID, title, body)
	}
	s.recordEvent(ctx, v, "admin", uuid.Nil, body)
}

func (s *Service) recordEvent(ctx context.Context, v *domain.Volunteer, role string, actor uuid.UUID, comment string) {
	_ = s.volunteers.AddEvent(ctx, &domain.VolunteerEvent{
		ID:          uuid.New(),
		VolunteerID: v.ID,
		ActorUserID: actor,
		ActorRole:   role,
		EventType:   domain.EventCertificate,
		FromStatus:  v.Status,
		ToStatus:    v.Status,
		Comment:     comment,
		CreatedAt:   s.clock.Now(),
	})
}

func (s *Service) Verify(ctx context.Context, code uuid.UUID) (*domain.Certificate, error) {
	c, err := s.certs.GetByVerificationCode(ctx, code)
	if err != nil {
		return nil, err
	}
	v, err := s.volunteers.GetByID(ctx, c.VolunteerID)
	if err == nil {
		c.Volunteer = v
	}
	return c, nil
}

func (s *Service) ListMine(ctx context.Context, userID uuid.UUID) ([]domain.Certificate, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.certs.ListByVolunteer(ctx, v.ID)
}

func (s *Service) ListByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]domain.Certificate, error) {
	return s.certs.ListByVolunteer(ctx, volunteerID)
}

func (s *Service) PDF(ctx context.Context, code uuid.UUID) ([]byte, *domain.Certificate, error) {
	c, err := s.Verify(ctx, code)
	if err != nil {
		return nil, nil, err
	}
	verifyURL := fmt.Sprintf("%s/verify/%s", s.publicBase, c.VerificationCode.String())
	codeImg, err := qr.Encode(verifyURL, qr.M, qr.Auto)
	if err != nil {
		return nil, nil, err
	}
	codeImg, err = barcode.Scale(codeImg, 256, 256)
	if err != nil {
		return nil, nil, err
	}
	rgba := image.NewNRGBA(codeImg.Bounds())
	draw.Draw(rgba, rgba.Bounds(), codeImg, codeImg.Bounds().Min, draw.Src)
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, rgba); err != nil {
		return nil, nil, err
	}
	pngBytes := pngBuf.Bytes()

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetDrawColor(225, 90, 41)
	pdf.SetLineWidth(2)
	pdf.Rect(10, 10, 277, 190, "D")
	pdf.SetLineWidth(0.4)
	pdf.Rect(14, 14, 269, 182, "D")

	pdf.SetTextColor(180, 180, 180)
	pdf.SetFont("Helvetica", "B", 48)
	pdf.TransformBegin()
	pdf.TransformRotate(30, 148, 105)
	pdf.Text(60, 120, "MAHAK - VOLUNTEER")
	pdf.TransformEnd()

	pdf.SetTextColor(225, 90, 41)
	pdf.SetFont("Helvetica", "B", 28)
	pdf.SetXY(20, 28)
	heading := "MAHAK Appreciation Letter"
	if c.Kind == domain.CertOfficial {
		heading = "MAHAK Volunteer Activity Certificate"
	}
	pdf.CellFormat(257, 14, heading, "", 1, "C", false, 0, "")

	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("Helvetica", "", 14)
	name := "Volunteer"
	if c.Volunteer != nil && isASCII(c.Volunteer.FullName) {
		name = c.Volunteer.FullName
	}
	pdf.SetXY(20, 55)
	pdf.MultiCell(200, 8, fmt.Sprintf("This certifies that %s has completed volunteer service for MAHAK (Society to Support Children Suffering from Cancer).", name), "", "L", false)

	pdf.SetXY(20, 85)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(40, 10, "Activity:")
	pdf.SetFont("Helvetica", "", 16)
	pdf.Cell(0, 10, asciiOrDash(c.Title))

	pdf.SetXY(20, 100)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(40, 10, "Hours:")
	pdf.SetFont("Helvetica", "", 16)
	pdf.Cell(0, 10, fmt.Sprintf("%.1f hours", c.Hours))

	pdf.SetXY(20, 115)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(40, 10, "Issued:")
	pdf.SetFont("Helvetica", "", 16)
	pdf.Cell(0, 10, c.IssuedAt.Format("2006-01-02"))

	pdf.SetXY(20, 130)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(40, 8, "UUID:")
	pdf.SetFont("Courier", "", 11)
	pdf.Cell(0, 8, c.VerificationCode.String())

	opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	pdf.RegisterImageOptionsReader("qr", opt, bytes.NewReader(pngBytes))
	pdf.ImageOptions("qr", 220, 70, 50, 50, false, opt, 0, "")
	pdf.SetXY(215, 122)
	pdf.SetFont("Helvetica", "", 8)
	pdf.MultiCell(60, 4, "Scan to verify authenticity", "", "C", false)

	pdf.SetXY(20, 170)
	pdf.SetFont("Helvetica", "I", 10)
	pdf.SetTextColor(90, 90, 90)
	pdf.Cell(0, 6, "Mahak - Digital Volunteer Module  |  Watermarked official record")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), c, nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func asciiOrDash(s string) string {
	if isASCII(s) && s != "" {
		return s
	}
	return "Mahak volunteer activity"
}
