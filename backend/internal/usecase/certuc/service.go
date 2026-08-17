package certuc

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/png"
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
	clock      domain.Clock
	publicBase string
}

func New(certs domain.CertificateRepository, tasks domain.TaskRepository, volunteers domain.VolunteerRepository, clock domain.Clock, publicBase string) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	if publicBase == "" {
		publicBase = "http://localhost:3000"
	}
	return &Service{certs: certs, tasks: tasks, volunteers: volunteers, clock: clock, publicBase: publicBase}
}

func (s *Service) IssueForAssignment(ctx context.Context, assignmentID uuid.UUID) (*domain.Certificate, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if !a.CanIssueCertificate() {
		return nil, domain.ErrCertificateNotReady
	}
	exists, err := s.certs.ExistsForAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrConflict
	}
	v, err := s.volunteers.GetByID(ctx, a.VolunteerID)
	if err != nil {
		return nil, err
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	c := &domain.Certificate{
		ID:               uuid.New(),
		VerificationCode: uuid.New(),
		VolunteerID:      v.ID,
		Kind:             domain.CertTask,
		AssignmentID:     &a.ID,
		Title:            t.Title,
		Hours:            a.HoursAwarded,
		IssuedAt:         now,
		Volunteer:        v,
	}
	return c, s.certs.Create(ctx, c)
}

func (s *Service) IssueAggregated(ctx context.Context, volunteerID uuid.UUID, from, to time.Time) (*domain.Certificate, error) {
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	if v.TotalHours <= 0 {
		return nil, domain.ErrCertificateNotReady
	}
	now := s.clock.Now()
	title := fmt.Sprintf("گواهی تجمیعی همکاری داوطلبانه")
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
	pdf.CellFormat(257, 14, "MAHAK Volunteer Certificate", "", 1, "C", false, 0, "")

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
