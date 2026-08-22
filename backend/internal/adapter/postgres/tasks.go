package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type TaskRepo struct{ db *DB }

func (d *DB) Tasks() *TaskRepo { return &TaskRepo{d} }

func (r *TaskRepo) Create(ctx context.Context, t *domain.Task) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO tasks (
		id,title,description,location,starts_at,ends_at,capacity,reserved_count,hour_weight,
		required_skills,required_skill_ids,min_score,required_education,work_mode,delivery_hint,status,created_by,created_at,updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		t.ID, t.Title, t.Description, t.Location, t.StartsAt, t.EndsAt, t.Capacity, t.ReservedCount,
		t.HourWeight, skillsToText(t.RequiredSkills), t.RequiredSkillIDs, t.MinScore, t.RequiredEducation,
		t.WorkMode, t.DeliveryHint, t.Status, nilUUID(t.CreatedBy), t.CreatedAt, t.UpdatedAt)
	return mapErr(err)
}

func (r *TaskRepo) Update(ctx context.Context, t *domain.Task) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE tasks SET title=$2,description=$3,location=$4,starts_at=$5,ends_at=$6,
		capacity=$7,reserved_count=$8,hour_weight=$9,required_skills=$10,required_skill_ids=$11,min_score=$12,required_education=$13,
		work_mode=$14,delivery_hint=$15,status=$16,updated_at=$17 WHERE id=$1`,
		t.ID, t.Title, t.Description, t.Location, t.StartsAt, t.EndsAt, t.Capacity, t.ReservedCount,
		t.HourWeight, skillsToText(t.RequiredSkills), t.RequiredSkillIDs, t.MinScore, t.RequiredEducation,
		t.WorkMode, t.DeliveryHint, t.Status, t.UpdatedAt)
	return mapErr(err)
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return scanTask(r.db.Pool.QueryRow(ctx, taskCols+` WHERE id=$1`, id))
}

func (r *TaskRepo) CloseExpired(ctx context.Context, now time.Time) (int64, error) {
	tag, err := r.db.Pool.Exec(ctx, `UPDATE tasks SET status='closed', updated_at=$1 WHERE status='open' AND ends_at < $1`, now)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *TaskRepo) List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, int, error) {
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", n))
		args = append(args, f.Status)
		n++
	}
	if f.Skill != "" {
		where = append(where, fmt.Sprintf("$%d = ANY(required_skills)", n))
		args = append(args, string(f.Skill))
		n++
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d OR location ILIKE $%d)", n, n, n))
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if f.Upcoming {
		where = append(where, "ends_at > now()")
	}
	if f.ExcludeVolunteerID != uuid.Nil {
		where = append(where, fmt.Sprintf(`id NOT IN (
			SELECT task_id FROM assignments
			WHERE volunteer_id=$%d AND status IN ('requested','reserved','in_progress','attended','submitted','completed')
		)`, n))
		args = append(args, f.ExcludeVolunteerID)
		n++
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE "+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	q := fmt.Sprintf(taskCols+` WHERE %s ORDER BY starts_at ASC LIMIT $%d OFFSET $%d`, w, n, n+1)
	args = append(args, limit, f.Offset)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *t)
	}
	return out, total, rows.Err()
}

func (r *TaskRepo) ApplySeat(ctx context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	existing, err := r.GetAssignmentByTaskVolunteer(ctx, taskID, volunteerID)
	if err == nil {
		if existing.Status.BlocksReapply() {
			return nil, domain.ErrAlreadyAssigned
		}
		existing.Status = domain.AssignmentRequested
		existing.AdminComment = ""
		if err := r.UpdateAssignment(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if err != domain.ErrNotFound {
		return nil, err
	}
	now := time.Now().UTC()
	a := &domain.Assignment{
		ID:          uuid.New(),
		TaskID:      taskID,
		VolunteerID: volunteerID,
		Status:      domain.AssignmentRequested,
		CreatedAt:   now,
	}
	_, err = r.db.Pool.Exec(ctx, `INSERT INTO assignments (id,task_id,volunteer_id,status,created_at) VALUES ($1,$2,$3,$4,$5)`,
		a.ID, a.TaskID, a.VolunteerID, a.Status, a.CreatedAt)
	if err != nil {
		if mapped := mapErr(err); mapped == domain.ErrConflict {
			return nil, domain.ErrAlreadyAssigned
		}
		return nil, mapErr(err)
	}
	return a, nil
}

func (r *TaskRepo) ReserveSeat(ctx context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var t domain.Task
	var skills []string
	err = tx.QueryRow(ctx, taskCols+` WHERE id=$1 FOR UPDATE`, taskID).Scan(
		&t.ID, &t.Title, &t.Description, &t.Location, &t.StartsAt, &t.EndsAt, &t.Capacity, &t.ReservedCount,
		&t.HourWeight, &skills, &t.MinScore, &t.RequiredEducation, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.RequiredSkillIDs,
		&t.WorkMode, &t.DeliveryHint,
	)
	if err != nil {
		return nil, mapErr(err)
	}
	if t.ReservedCount >= t.Capacity {
		return nil, domain.ErrCapacityFull
	}
	now := time.Now().UTC()
	a := &domain.Assignment{
		ID:          uuid.New(),
		TaskID:      taskID,
		VolunteerID: volunteerID,
		Status:      domain.AssignmentReserved,
		CreatedAt:   now,
	}
	_, err = tx.Exec(ctx, `INSERT INTO assignments (id,task_id,volunteer_id,status,created_at) VALUES ($1,$2,$3,$4,$5)`,
		a.ID, a.TaskID, a.VolunteerID, a.Status, a.CreatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE tasks SET reserved_count = reserved_count + 1, updated_at=$2 WHERE id=$1`, taskID, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *TaskRepo) GetAssignment(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	return scanAssignment(r.db.Pool.QueryRow(ctx, assignmentCols+` WHERE a.id=$1`, id))
}

func (r *TaskRepo) GetAssignmentByTaskVolunteer(ctx context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	return scanAssignment(r.db.Pool.QueryRow(ctx, assignmentCols+` WHERE a.task_id=$1 AND a.volunteer_id=$2`, taskID, volunteerID))
}

func (r *TaskRepo) UpdateAssignment(ctx context.Context, a *domain.Assignment) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE assignments SET status=$2,volunteer_rating=$3,volunteer_comment=$4,
		admin_discipline=$5,admin_expertise=$6,admin_ethics=$7,admin_comment=$8,composite_score=$9,
		hours_awarded=$10,attended_at=$11,completed_at=$12,delivery_note=$13,delivery_file_name=$14,
		delivery_object_key=$15,delivery_mime=$16,delivered_at=$17 WHERE id=$1`,
		a.ID, a.Status, a.VolunteerRating, a.VolunteerComment, a.AdminDiscipline, a.AdminExpertise,
		a.AdminEthics, a.AdminComment, a.CompositeScore, a.HoursAwarded, a.AttendedAt, a.CompletedAt,
		a.DeliveryNote, a.DeliveryFileName, a.DeliveryObjectKey, a.DeliveryMime, a.DeliveredAt)
	return mapErr(err)
}

func (r *TaskRepo) ListAssignments(ctx context.Context, f domain.AssignmentFilter) ([]domain.Assignment, int, error) {
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if f.VolunteerID != uuid.Nil {
		where = append(where, fmt.Sprintf("a.volunteer_id=$%d", n))
		args = append(args, f.VolunteerID)
		n++
	}
	if f.TaskID != uuid.Nil {
		where = append(where, fmt.Sprintf("a.task_id=$%d", n))
		args = append(args, f.TaskID)
		n++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("a.status=$%d", n))
		args = append(args, f.Status)
		n++
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM assignments a WHERE "+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	q := fmt.Sprintf(assignmentCols+` WHERE %s ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d`, w, n, n+1)
	args = append(args, limit, f.Offset)
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Assignment
	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, rows.Err()
}

const taskCols = `SELECT id,title,description,COALESCE(location,''),starts_at,ends_at,capacity,reserved_count,hour_weight,
	required_skills,min_score,COALESCE(required_education,''),status,COALESCE(created_by,'00000000-0000-0000-0000-000000000000'),created_at,updated_at,
	COALESCE(required_skill_ids, '{}'), COALESCE(work_mode,'onsite'), COALESCE(delivery_hint,'') FROM tasks`

func scanTask(row pgx.Row) (*domain.Task, error) {
	var t domain.Task
	var skills []string
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Location, &t.StartsAt, &t.EndsAt, &t.Capacity, &t.ReservedCount,
		&t.HourWeight, &skills, &t.MinScore, &t.RequiredEducation, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.RequiredSkillIDs,
		&t.WorkMode, &t.DeliveryHint)
	if err != nil {
		return nil, mapErr(err)
	}
	t.RequiredSkills = domain.ParseSkillCategories(skills)
	if t.RequiredSkillIDs == nil {
		t.RequiredSkillIDs = []uuid.UUID{}
	}
	if t.WorkMode == "" {
		t.WorkMode = domain.WorkOnsite
	}
	return &t, nil
}

const assignmentCols = `SELECT a.id,a.task_id,a.volunteer_id,a.status,a.volunteer_rating,COALESCE(a.volunteer_comment,''),
	a.admin_discipline,a.admin_expertise,a.admin_ethics,COALESCE(a.admin_comment,''),a.composite_score,a.hours_awarded,
	a.attended_at,a.completed_at,a.created_at,
	COALESCE(a.delivery_note,''), COALESCE(a.delivery_file_name,''), COALESCE(a.delivery_object_key,''), COALESCE(a.delivery_mime,''), a.delivered_at,
	t.title, t.hour_weight, COALESCE(t.location,''), t.starts_at, t.ends_at, COALESCE(t.work_mode,'onsite'), COALESCE(t.delivery_hint,''),
	v.full_name, COALESCE(v.phone,'')
	FROM assignments a
	JOIN tasks t ON t.id=a.task_id
	JOIN volunteers v ON v.id=a.volunteer_id`

func scanAssignment(row pgx.Row) (*domain.Assignment, error) {
	var a domain.Assignment
	a.Task = &domain.Task{}
	a.Volunteer = &domain.Volunteer{}
	err := row.Scan(&a.ID, &a.TaskID, &a.VolunteerID, &a.Status, &a.VolunteerRating, &a.VolunteerComment,
		&a.AdminDiscipline, &a.AdminExpertise, &a.AdminEthics, &a.AdminComment, &a.CompositeScore, &a.HoursAwarded,
		&a.AttendedAt, &a.CompletedAt, &a.CreatedAt,
		&a.DeliveryNote, &a.DeliveryFileName, &a.DeliveryObjectKey, &a.DeliveryMime, &a.DeliveredAt,
		&a.Task.Title, &a.Task.HourWeight, &a.Task.Location, &a.Task.StartsAt, &a.Task.EndsAt, &a.Task.WorkMode, &a.Task.DeliveryHint,
		&a.Volunteer.FullName, &a.Volunteer.Phone)
	if err != nil {
		return nil, mapErr(err)
	}
	a.Task.ID = a.TaskID
	a.Volunteer.ID = a.VolunteerID
	return &a, nil
}

func nilUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
