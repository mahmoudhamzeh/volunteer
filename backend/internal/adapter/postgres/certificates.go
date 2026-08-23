package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

const certReqCols = `SELECT r.id, r.volunteer_id, r.kind, r.assignment_id, r.status, COALESCE(r.admin_note,''), r.certificate_id, r.created_at, r.reviewed_at,
	v.full_name, COALESCE(t.title, '')
	FROM certificate_requests r
	JOIN volunteers v ON v.id = r.volunteer_id
	LEFT JOIN assignments a ON a.id = r.assignment_id
	LEFT JOIN tasks t ON t.id = a.task_id`

func (r *CertRepo) CreateRequest(ctx context.Context, req *domain.CertificateRequest) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO certificate_requests (id, volunteer_id, kind, assignment_id, status, admin_note, certificate_id, created_at, reviewed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		req.ID, req.VolunteerID, req.Kind, req.AssignmentID, req.Status, req.AdminNote, req.CertificateID, req.CreatedAt, req.ReviewedAt)
	return mapErr(err)
}

func (r *CertRepo) GetRequest(ctx context.Context, id uuid.UUID) (*domain.CertificateRequest, error) {
	return scanCertReq(r.db.Pool.QueryRow(ctx, certReqCols+` WHERE r.id=$1`, id))
}

func (r *CertRepo) UpdateRequest(ctx context.Context, req *domain.CertificateRequest) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE certificate_requests SET status=$2, admin_note=$3, certificate_id=$4, reviewed_at=$5 WHERE id=$1`,
		req.ID, req.Status, req.AdminNote, req.CertificateID, req.ReviewedAt)
	return mapErr(err)
}

func (r *CertRepo) ListRequests(ctx context.Context, status domain.CertificateRequestStatus) ([]domain.CertificateRequest, error) {
	q := certReqCols + ` ORDER BY r.created_at DESC`
	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.db.Pool.Query(ctx, certReqCols+` WHERE r.status=$1 ORDER BY r.created_at DESC`, status)
	} else {
		rows, err = r.db.Pool.Query(ctx, q)
	}
	if err != nil {
		return nil, err
	}
	return collectCertReqs(rows)
}

func (r *CertRepo) ListRequestsByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]domain.CertificateRequest, error) {
	rows, err := r.db.Pool.Query(ctx, certReqCols+` WHERE r.volunteer_id=$1 ORDER BY r.created_at DESC`, volunteerID)
	if err != nil {
		return nil, err
	}
	return collectCertReqs(rows)
}

func (r *CertRepo) HasPendingRequest(ctx context.Context, volunteerID uuid.UUID, kind domain.CertificateKind, assignmentID *uuid.UUID) (bool, error) {
	var n int
	var err error
	if assignmentID != nil {
		err = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM certificate_requests WHERE volunteer_id=$1 AND kind=$2 AND assignment_id=$3 AND status='pending'`,
			volunteerID, kind, *assignmentID).Scan(&n)
	} else {
		err = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM certificate_requests WHERE volunteer_id=$1 AND kind=$2 AND assignment_id IS NULL AND status='pending'`,
			volunteerID, kind).Scan(&n)
	}
	return n > 0, err
}

func collectCertReqs(rows pgx.Rows) ([]domain.CertificateRequest, error) {
	defer rows.Close()
	var out []domain.CertificateRequest
	for rows.Next() {
		req, err := scanCertReq(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *req)
	}
	return out, rows.Err()
}

func scanCertReq(row pgx.Row) (*domain.CertificateRequest, error) {
	var req domain.CertificateRequest
	err := row.Scan(&req.ID, &req.VolunteerID, &req.Kind, &req.AssignmentID, &req.Status, &req.AdminNote, &req.CertificateID, &req.CreatedAt, &req.ReviewedAt,
		&req.VolunteerName, &req.AssignmentTitle)
	if err != nil {
		return nil, mapErr(err)
	}
	return &req, nil
}
