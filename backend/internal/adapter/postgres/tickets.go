package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type TicketRepo struct{ db *DB }

func (d *DB) Tickets() *TicketRepo { return &TicketRepo{d} }

func (r *TicketRepo) Create(ctx context.Context, t *domain.Ticket) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO tickets (id, volunteer_id, subject, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		t.ID, t.VolunteerID, t.Subject, t.Status, t.CreatedAt, t.UpdatedAt)
	return mapErr(err)
}

func (r *TicketRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	t, err := scanTicket(r.db.Pool.QueryRow(ctx, ticketCols+` WHERE t.id=$1`, id))
	if err != nil {
		return nil, err
	}
	msgs, err := r.ListMessages(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Messages = msgs
	return t, nil
}

func (r *TicketRepo) Update(ctx context.Context, t *domain.Ticket) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE tickets SET subject=$2, status=$3, updated_at=$4 WHERE id=$1`,
		t.ID, t.Subject, t.Status, t.UpdatedAt)
	return mapErr(err)
}

func (r *TicketRepo) ListByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]domain.Ticket, error) {
	rows, err := r.db.Pool.Query(ctx, ticketCols+` WHERE t.volunteer_id=$1 ORDER BY t.updated_at DESC`, volunteerID)
	if err != nil {
		return nil, err
	}
	return collectTickets(rows)
}

func (r *TicketRepo) List(ctx context.Context, status domain.TicketStatus) ([]domain.Ticket, error) {
	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.db.Pool.Query(ctx, ticketCols+` WHERE t.status=$1 ORDER BY t.updated_at DESC`, status)
	} else {
		rows, err = r.db.Pool.Query(ctx, ticketCols+` ORDER BY CASE t.status WHEN 'open' THEN 0 WHEN 'answered' THEN 1 ELSE 2 END, t.updated_at DESC`)
	}
	if err != nil {
		return nil, err
	}
	return collectTickets(rows)
}

func (r *TicketRepo) AddMessage(ctx context.Context, m *domain.TicketMessage) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO ticket_messages (id, ticket_id, author_user_id, author_role, body, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		m.ID, m.TicketID, nilUUID(m.AuthorUserID), m.AuthorRole, m.Body, m.CreatedAt)
	return mapErr(err)
}

func (r *TicketRepo) ListMessages(ctx context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id, ticket_id, COALESCE(author_user_id, '00000000-0000-0000-0000-000000000000'), author_role, body, created_at
		FROM ticket_messages WHERE ticket_id=$1 ORDER BY created_at`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TicketMessage
	for rows.Next() {
		var m domain.TicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.AuthorUserID, &m.AuthorRole, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

const ticketCols = `SELECT t.id, t.volunteer_id, t.subject, t.status, t.created_at, t.updated_at, v.full_name
	FROM tickets t JOIN volunteers v ON v.id = t.volunteer_id`

func collectTickets(rows pgx.Rows) ([]domain.Ticket, error) {
	defer rows.Close()
	var out []domain.Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func scanTicket(row pgx.Row) (*domain.Ticket, error) {
	var t domain.Ticket
	err := row.Scan(&t.ID, &t.VolunteerID, &t.Subject, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.VolunteerName)
	if err != nil {
		return nil, mapErr(err)
	}
	return &t, nil
}
