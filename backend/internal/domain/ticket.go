package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TicketStatus string

const (
	TicketOpen     TicketStatus = "open"
	TicketAnswered TicketStatus = "answered"
	TicketClosed   TicketStatus = "closed"
)

var TicketSubjects = []string{
	"سوال درباره فعالیت",
	"آموزش",
	"حضور و غیاب",
	"مدارک و پرونده",
	"تقدیرنامه و گواهی",
	"زمان‌بندی و انصراف",
	"پیشنهاد و انتقاد",
	"سایر",
}

func ValidTicketSubject(subject string) bool {
	for _, s := range TicketSubjects {
		if s == subject {
			return true
		}
	}
	return false
}

type Ticket struct {
	ID            uuid.UUID       `json:"id"`
	VolunteerID   uuid.UUID       `json:"volunteer_id"`
	VolunteerName string          `json:"volunteer_name,omitempty"`
	Subject       string          `json:"subject"`
	Status        TicketStatus    `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Messages      []TicketMessage `json:"messages,omitempty"`
}

type TicketMessage struct {
	ID           uuid.UUID `json:"id"`
	TicketID     uuid.UUID `json:"ticket_id"`
	AuthorUserID uuid.UUID `json:"author_user_id"`
	AuthorRole   string    `json:"author_role"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
}

type TicketRepository interface {
	Create(ctx context.Context, t *Ticket) error
	Get(ctx context.Context, id uuid.UUID) (*Ticket, error)
	Update(ctx context.Context, t *Ticket) error
	ListByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]Ticket, error)
	List(ctx context.Context, status TicketStatus) ([]Ticket, error)
	AddMessage(ctx context.Context, m *TicketMessage) error
	ListMessages(ctx context.Context, ticketID uuid.UUID) ([]TicketMessage, error)
}
