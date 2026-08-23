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
		id,user_id,full_name,first_name,last_name,national_id,phone,phone2,province,city,address,plaque,unit,bio,skill_categories,
		education_level,education_field,medical_license,birth_date,status,rejection_reason,average_score,total_hours,completed_tasks,created_at,updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,NULLIF($19,'')::date,$20,$21,$22,$23,$24,$25,$26)`,
		v.ID, v.UserID, v.FullName, v.FirstName, v.LastName, v.NationalID, v.Phone, v.Phone2, v.Province, v.City, v.Address, v.Plaque, v.Unit, v.Bio,
		skillsToText(v.SkillCategories), v.EducationLevel, v.EducationField, v.MedicalLicense, v.BirthDate, v.Status, v.RejectionReason,
		v.AverageScore, v.TotalHours, v.CompletedTasks, v.CreatedAt, v.UpdatedAt)
	return mapErr(err)
}

func (r *VolunteerRepo) Update(ctx context.Context, v *domain.Volunteer) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE volunteers SET
		full_name=$2,first_name=$3,last_name=$4,national_id=$5,phone=$6,phone2=$7,province=$8,city=$9,address=$10,plaque=$11,unit=$12,bio=$13,
		skill_categories=$14,education_level=$15,education_field=$16,medical_license=$17,birth_date=NULLIF($18,'')::date,status=$19,rejection_reason=$20,
		average_score=$21,total_hours=$22,completed_tasks=$23,updated_at=$24 WHERE id=$1`,
		v.ID, v.FullName, v.FirstName, v.LastName, v.NationalID, v.Phone, v.Phone2, v.Province, v.City, v.Address, v.Plaque, v.Unit, v.Bio,
		skillsToText(v.SkillCategories), v.EducationLevel, v.EducationField, v.MedicalLicense, v.BirthDate, v.Status, v.RejectionReason,
		v.AverageScore, v.TotalHours, v.CompletedTasks, v.UpdatedAt)
	return mapErr(err)
}

func (r *VolunteerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volunteer, error) {
	return scanVolunteer(r.db.Pool.QueryRow(ctx, volunteerSelect+` WHERE v.id=$1`, id))
}

func (r *VolunteerRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Volunteer, error) {
	return scanVolunteer(r.db.Pool.QueryRow(ctx, volunteerSelect+` WHERE v.user_id=$1`, userID))
}

func (r *VolunteerRepo) GetByPhone(ctx context.Context, phone string) (*domain.Volunteer, error) {
	return scanVolunteer(r.db.Pool.QueryRow(ctx, volunteerSelect+` WHERE v.phone=$1 AND v.phone <> '' LIMIT 1`, phone))
}

func (r *VolunteerRepo) List(ctx context.Context, f domain.VolunteerFilter) ([]domain.Volunteer, int, error) {
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if f.Status != "" {
		where = append(where, fmt.Sprintf("v.status=$%d", n))
		args = append(args, f.Status)
		n++
	}
	if f.Skill != "" {
		where = append(where, fmt.Sprintf("$%d = ANY(v.skill_categories)", n))
		args = append(args, string(f.Skill))
		n++
	}
	if f.Attention == "resubmitted" {
		where = append(where, resubmittedDocsSQL)
		if f.Status == "" {
			// status already constrained inside resubmittedDocsSQL
		}
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(v.full_name ILIKE $%d OR v.national_id ILIKE $%d OR v.phone ILIKE $%d OR v.city ILIKE $%d OR v.province ILIKE $%d OR u.email ILIKE $%d)", n, n, n, n, n, n))
		args = append(args, "%"+f.Query+"%")
		n++
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM volunteers v LEFT JOIN users u ON u.id = v.user_id WHERE "+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := f.Limit, f.Offset
	if limit <= 0 {
		limit = 20
	}
	q := fmt.Sprintf(volunteerSelect+` WHERE %s ORDER BY v.created_at DESC LIMIT $%d OFFSET $%d`, w, n, n+1)
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

func (r *VolunteerRepo) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM documents WHERE id=$1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *VolunteerRepo) AddEvent(ctx context.Context, e *domain.VolunteerEvent) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO volunteer_events
		(id, volunteer_id, actor_user_id, actor_role, event_type, from_status, to_status, comment, created_at)
		VALUES ($1,$2,NULLIF($3,'00000000-0000-0000-0000-000000000000')::uuid,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.VolunteerID, e.ActorUserID, e.ActorRole, e.EventType, e.FromStatus, e.ToStatus, e.Comment, e.CreatedAt)
	return mapErr(err)
}

func (r *VolunteerRepo) ListEvents(ctx context.Context, volunteerID uuid.UUID, limit int) ([]domain.VolunteerEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := r.db.Pool.Query(ctx, `SELECT id, volunteer_id, COALESCE(actor_user_id, '00000000-0000-0000-0000-000000000000'),
		COALESCE(actor_role,''), event_type, COALESCE(from_status,''), COALESCE(to_status,''), COALESCE(comment,''), created_at
		FROM volunteer_events WHERE volunteer_id=$1 ORDER BY created_at DESC LIMIT $2`, volunteerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.VolunteerEvent
	for rows.Next() {
		var e domain.VolunteerEvent
		if err := rows.Scan(&e.ID, &e.VolunteerID, &e.ActorUserID, &e.ActorRole, &e.EventType, &e.FromStatus, &e.ToStatus, &e.Comment, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if out == nil {
		out = []domain.VolunteerEvent{}
	}
	return out, rows.Err()
}

func (r *VolunteerRepo) ReplaceSkills(ctx context.Context, volunteerID uuid.UUID, skillIDs []uuid.UUID) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM volunteer_skills WHERE volunteer_id=$1`, volunteerID); err != nil {
		return err
	}
	for _, id := range skillIDs {
		if id == uuid.Nil {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO volunteer_skills (volunteer_id, skill_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, volunteerID, id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *VolunteerRepo) ListVolunteerSkills(ctx context.Context, volunteerID uuid.UUID) ([]domain.VolunteerSkill, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT s.id, s.title, g.id, g.slug, g.title
		FROM volunteer_skills vs
		JOIN skills s ON s.id = vs.skill_id
		JOIN skill_groups g ON g.id = s.group_id
		WHERE vs.volunteer_id=$1
		ORDER BY g.sort_order, s.title`, volunteerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.VolunteerSkill
	for rows.Next() {
		var s domain.VolunteerSkill
		if err := rows.Scan(&s.SkillID, &s.Title, &s.GroupID, &s.GroupSlug, &s.GroupTitle); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []domain.VolunteerSkill{}
	}
	return out, rows.Err()
}

const resubmittedDocsSQL = `v.status IN ('pending','draft','rejected')
	AND EXISTS (
		SELECT 1 FROM volunteer_events e
		WHERE e.volunteer_id = v.id AND e.event_type IN ('documents_requested','rejected')
	)
	AND EXISTS (
		SELECT 1 FROM documents d
		WHERE d.volunteer_id = v.id
		AND d.created_at > (
			SELECT MAX(e.created_at) FROM volunteer_events e
			WHERE e.volunteer_id = v.id AND e.event_type IN ('documents_requested','rejected')
		)
	)`

const volunteerSelect = `SELECT v.id,v.user_id,v.full_name,COALESCE(v.first_name,''),COALESCE(v.last_name,''),COALESCE(v.national_id,''),COALESCE(v.phone,''),COALESCE(v.phone2,''),
	COALESCE(v.province,''),COALESCE(v.city,''),COALESCE(v.address,''),COALESCE(v.plaque,''),COALESCE(v.unit,''),COALESCE(v.bio,''),v.skill_categories,
	COALESCE(v.education_level,''),COALESCE(v.education_field,''),COALESCE(v.medical_license,''),COALESCE(to_char(v.birth_date,'YYYY-MM-DD'),''),
	v.status,COALESCE(v.rejection_reason,''),v.average_score,v.total_hours,v.completed_tasks,v.created_at,v.updated_at,COALESCE(u.email,'')
	FROM volunteers v LEFT JOIN users u ON u.id = v.user_id`

func scanVolunteer(row pgx.Row) (*domain.Volunteer, error) {
	var v domain.Volunteer
	var skills []string
	err := row.Scan(&v.ID, &v.UserID, &v.FullName, &v.FirstName, &v.LastName, &v.NationalID, &v.Phone, &v.Phone2, &v.Province, &v.City, &v.Address,
		&v.Plaque, &v.Unit, &v.Bio, &skills, &v.EducationLevel, &v.EducationField, &v.MedicalLicense, &v.BirthDate, &v.Status,
		&v.RejectionReason, &v.AverageScore, &v.TotalHours, &v.CompletedTasks, &v.CreatedAt, &v.UpdatedAt, &v.Email)
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
