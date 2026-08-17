package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type VolunteerRepo struct{ db *DB }

func (d *DB) Volunteers() *VolunteerRepo { return &VolunteerRepo{d} }

func (r *VolunteerRepo) Create(ctx context.Context, v *domain.Volunteer) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO volunteers (
		id,user_id,full_name,national_id,phone,city,bio,skill_categories,education_field,medical_license,
		status,rejection_reason,average_score,total_hours,completed_tasks,created_at,updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		v.ID, v.UserID, v.FullName, v.NationalID, v.Phone, v.City, v.Bio, skillsToText(v.SkillCategories),
		v.EducationField, v.MedicalLicense, v.Status, v.RejectionReason, v.AverageScore, v.TotalHours,
		v.CompletedTasks, v.CreatedAt, v.UpdatedAt)
	return mapErr(err)
}

func (r *VolunteerRepo) Update(ctx context.Context, v *domain.Volunteer) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE volunteers SET
		full_name=$2,national_id=$3,phone=$4,city=$5,bio=$6,skill_categories=$7,education_field=$8,
		medical_license=$9,status=$10,rejection_reason=$11,average_score=$12,total_hours=$13,
		completed_tasks=$14,updated_at=$15 WHERE id=$1`,
		v.ID, v.FullName, v.NationalID, v.Phone, v.City, v.Bio, skillsToText(v.SkillCategories),
		v.EducationField, v.MedicalLicense, v.Status, v.RejectionReason, v.AverageScore, v.TotalHours,
		v.CompletedTasks, v.UpdatedAt)
	return mapErr(err)
}

func (r *VolunteerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volunteer, error) {
	return scanVolunteer(r.db.Pool.QueryRow(ctx, volunteerCols+` WHERE id=$1`, id))
}

func (r *VolunteerRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	return scanVolunteer(r.db.Pool.QueryRow(ctx, volunteerCols+` WHERE user_id=$1`, userID))
}

func (r *VolunteerRepo) List(ctx context.Context, f domain.VolunteerFilter) ([]domain.Volunteer, int, error) {
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", n))
		args = append(args, f.Status)
		n++
	}
	if f.Skill != "" {
		where = append(where, fmt.Sprintf("$%d = ANY(skill_categories)", n))
		args = append(args, string(f.Skill))
		n++
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(full_name ILIKE $%d OR national_id ILIKE $%d OR phone ILIKE $%d)", n, n, n))
		args = append(args, "%"+f.Query+"%")
		n++
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM volunteers WHERE "+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := f.Limit, f.Offset
	if limit <= 0 {
		limit = 20
	}
	q := fmt.Sprintf(volunteerCols+` WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, w, n, n+1)
	args = append(args, limit, offset)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Volunteer
	for rows.Next() {
		v, err := scanVolunteer(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *v)
	}
	return out, total, rows.Err()
}

func (r *VolunteerRepo) ReplaceAvailability(ctx context.Context, volunteerID uuid.UUID, slots []domain.AvailabilitySlot) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM volunteer_availability WHERE volunteer_id=$1`, volunteerID); err != nil {
		return err
	}
	for _, sl := range slots {
		if _, err := tx.Exec(ctx, `INSERT INTO volunteer_availability (id,volunteer_id,weekday,start_time,end_time) VALUES ($1,$2,$3,$4,$5)`,
			sl.ID, volunteerID, sl.Weekday, sl.StartTime, sl.EndTime); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *VolunteerRepo) ListAvailability(ctx context.Context, volunteerID uuid.UUID) ([]domain.AvailabilitySlot, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id,volunteer_id,weekday,start_time,end_time FROM volunteer_availability WHERE volunteer_id=$1 ORDER BY weekday,start_time`, volunteerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AvailabilitySlot
	for rows.Next() {
		var s domain.AvailabilitySlot
		if err := rows.Scan(&s.ID, &s.VolunteerID, &s.Weekday, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *VolunteerRepo) AddDocument(ctx context.Context, d *domain.Document) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO documents (id,volunteer_id,kind,object_key,file_name,mime_type,size_bytes,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, d.ID, d.VolunteerID, d.Kind, d.ObjectKey, d.FileName, d.MimeType, d.SizeBytes, d.CreatedAt)
	return mapErr(err)
}

func (r *VolunteerRepo) ListDocuments(ctx context.Context, volunteerID uuid.UUID) ([]domain.Document, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id,volunteer_id,kind,object_key,file_name,mime_type,size_bytes,created_at FROM documents WHERE volunteer_id=$1 ORDER BY created_at DESC`, volunteerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Document
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.VolunteerID, &d.Kind, &d.ObjectKey, &d.FileName, &d.MimeType, &d.SizeBytes, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *VolunteerRepo) GetDocument(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	var d domain.Document
	err := r.db.Pool.QueryRow(ctx, `SELECT id,volunteer_id,kind,object_key,file_name,mime_type,size_bytes,created_at FROM documents WHERE id=$1`, id).
		Scan(&d.ID, &d.VolunteerID, &d.Kind, &d.ObjectKey, &d.FileName, &d.MimeType, &d.SizeBytes, &d.CreatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &d, nil
}

const volunteerCols = `SELECT id,user_id,full_name,COALESCE(national_id,''),COALESCE(phone,''),COALESCE(city,''),COALESCE(bio,''),skill_categories,
	COALESCE(education_field,''),COALESCE(medical_license,''),status,COALESCE(rejection_reason,''),average_score,total_hours,completed_tasks,created_at,updated_at FROM volunteers`

func scanVolunteer(row pgx.Row) (*domain.Volunteer, error) {
	var v domain.Volunteer
	var skills []string
	err := row.Scan(&v.ID, &v.UserID, &v.FullName, &v.NationalID, &v.Phone, &v.City, &v.Bio, &skills,
		&v.EducationField, &v.MedicalLicense, &v.Status, &v.RejectionReason, &v.AverageScore, &v.TotalHours,
		&v.CompletedTasks, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	v.SkillCategories = domain.ParseSkillCategories(skills)
	return &v, nil
}

func skillsToText(in []domain.SkillCategory) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = string(s)
	}
	return out
}
