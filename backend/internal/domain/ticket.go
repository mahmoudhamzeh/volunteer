package domain

import (
	"context"
	"strings"
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

func NormalizeTicketQuery(q string) (text, digits string) {
	q = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(q), "#"))
	var textB, digitB strings.Builder
	for _, r := range q {
		switch {
		case r >= '۰' && r <= '۹':
			n := rune('0' + (r - '۰'))
			textB.WriteRune(n)
			digitB.WriteRune(n)
		case r >= '٠' && r <= '٩':
			n := rune('0' + (r - '٠'))
			textB.WriteRune(n)
			digitB.WriteRune(n)
		default:
			textB.WriteRune(r)
			if r >= '0' && r <= '9' {
				digitB.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(textB.String()), digitB.String()
}

type Ticket struct {
	ID             uuid.UUID       `json:"id"`
	Number         int             `json:"number"`
	VolunteerID    uuid.UUID       `json:"volunteer_id"`
	VolunteerName  string          `json:"volunteer_name,omitempty"`
	VolunteerPhone string          `json:"volunteer_phone,omitempty"`
	Subject        string          `json:"subject"`
	Status         TicketStatus    `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Messages       []TicketMessage `json:"messages,omitempty"`
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
	List(ctx context.Context, status TicketStatus, q string) ([]Ticket, error)
	AddMessage(ctx context.Context, m *TicketMessage) error
	ListMessages(ctx context.Context, ticketID uuid.UUID) ([]TicketMessage, error)
}
