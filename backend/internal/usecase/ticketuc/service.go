package ticketuc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type staffNotifier interface {
	NotifyStaff(ctx context.Context, title, body string) error
}

type Service struct {
	tickets    domain.TicketRepository
	volunteers domain.VolunteerRepository
	notify     domain.Notifier
	clock      domain.Clock
}

func New(tickets domain.TicketRepository, volunteers domain.VolunteerRepository, notify domain.Notifier, clock domain.Clock) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	return &Service{tickets: tickets, volunteers: volunteers, notify: notify, clock: clock}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, subject, body string) (*domain.Ticket, error) {
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)
	if subject == "" {
		return nil, domain.Invalid("موضوع تیکت را بنویسید")
	}
	if body == "" {
		return nil, domain.Invalid("متن پرسش را بنویسید")
	}
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	t := &domain.Ticket{
		ID:            uuid.New(),
		VolunteerID:   v.ID,
		VolunteerName: v.FullName,
		Subject:       subject,
		Status:        domain.TicketOpen,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.tickets.Create(ctx, t); err != nil {
		return nil, err
	}
	msg := &domain.TicketMessage{
		ID:           uuid.New(),
		TicketID:     t.ID,
		AuthorUserID: userID,
		AuthorRole:   "volunteer",
		Body:         body,
		CreatedAt:    now,
	}
	if err := s.tickets.AddMessage(ctx, msg); err != nil {
		return nil, err
	}
	t.Messages = []domain.TicketMessage{*msg}
	if sn, ok := s.notify.(staffNotifier); ok {
		_ = sn.NotifyStaff(ctx, "تیکت جدید از داوطلب", v.FullName+" — "+subject)
	}
	return t, nil
}

func (s *Service) Mine(ctx context.Context, userID uuid.UUID) ([]domain.Ticket, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.tickets.ListByVolunteer(ctx, v.ID)
}

func (s *Service) GetMine(ctx context.Context, userID, ticketID uuid.UUID) (*domain.Ticket, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	t, err := s.tickets.Get(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if t.VolunteerID != v.ID {
		return nil, domain.ErrForbidden
	}
	return t, nil
}

func (s *Service) ReplyMine(ctx context.Context, userID, ticketID uuid.UUID, body string) (*domain.Ticket, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, domain.Invalid("متن پیام را بنویسید")
	}
	t, err := s.GetMine(ctx, userID, ticketID)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.TicketClosed {
		return nil, domain.Invalid("این تیکت بسته شده است و امکان ارسال پیام وجود ندارد")
	}
	now := s.clock.Now()
	if err := s.tickets.AddMessage(ctx, &domain.TicketMessage{
		ID: uuid.New(), TicketID: t.ID, AuthorUserID: userID, AuthorRole: "volunteer", Body: body, CreatedAt: now,
	}); err != nil {
		return nil, err
	}
	t.Status = domain.TicketOpen
	t.UpdatedAt = now
	if err := s.tickets.Update(ctx, t); err != nil {
		return nil, err
	}
	if sn, ok := s.notify.(staffNotifier); ok {
		_ = sn.NotifyStaff(ctx, "پاسخ داوطلب در تیکت", t.Subject)
	}
	return s.tickets.Get(ctx, t.ID)
}

func (s *Service) List(ctx context.Context, status domain.TicketStatus) ([]domain.Ticket, error) {
	return s.tickets.List(ctx, status)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	return s.tickets.Get(ctx, id)
}

func (s *Service) ReplyAdmin(ctx context.Context, actorID, ticketID uuid.UUID, body string) (*domain.Ticket, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, domain.Invalid("متن پاسخ را بنویسید")
	}
	t, err := s.tickets.Get(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.TicketClosed {
		return nil, domain.Invalid("این تیکت بسته شده است و امکان ارسال پیام وجود ندارد")
	}
	now := s.clock.Now()
	if err := s.tickets.AddMessage(ctx, &domain.TicketMessage{
		ID: uuid.New(), TicketID: t.ID, AuthorUserID: actorID, AuthorRole: "admin", Body: body, CreatedAt: now,
	}); err != nil {
		return nil, err
	}
	t.Status = domain.TicketAnswered
	t.UpdatedAt = now
	if err := s.tickets.Update(ctx, t); err != nil {
		return nil, err
	}
	v, err := s.volunteers.GetByID(ctx, t.VolunteerID)
	if err == nil && s.notify != nil {
		_ = s.notify.Notify(ctx, v.UserID, "پاسخ تیکت شما", "پشتیبانی به تیکت «"+t.Subject+"» پاسخ داد.")
	}
	return s.tickets.Get(ctx, t.ID)
}

func (s *Service) SetStatus(ctx context.Context, ticketID uuid.UUID, status domain.TicketStatus) (*domain.Ticket, error) {
	if status != domain.TicketOpen && status != domain.TicketAnswered && status != domain.TicketClosed {
		return nil, domain.Invalid("وضعیت تیکت نامعتبر است")
	}
	t, err := s.tickets.Get(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	t.Status = status
	t.UpdatedAt = s.clock.Now()
	if err := s.tickets.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
