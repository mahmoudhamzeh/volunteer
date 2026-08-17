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
		required_skills,required_skill_ids,min_score,required_education,status,created_by,created_at,updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		t.ID, t.Title, t.Description, t.Location, t.StartsAt, t.EndsAt, t.Capacity, t.ReservedCount,
		t.HourWeight, skillsToText(t.RequiredSkills), t.RequiredSkillIDs, t.MinScore, t.RequiredEducation, t.Status,
		nilUUID(t.CreatedBy), t.CreatedAt, t.UpdatedAt)
	return mapErr(err)
}

func (r *TaskRepo) Update(ctx context.Context, t *domain.Task) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE tasks SET title=$2,description=$3,location=$4,starts_at=$5,ends_at=$6,
		capacity=$7,reserved_count=$8,hour_weight=$9,required_skills=$10,required_skill_ids=$11,min_score=$12,required_education=$13,
		status=$14,updated_at=$15 WHERE id=$1`,
		t.ID, t.Title, t.Description, t.Location, t.StartsAt, t.EndsAt, t.Capacity, t.ReservedCount,
		t.HourWeight, skillsToText(t.RequiredSkills), t.RequiredSkillIDs, t.MinScore, t.RequiredEducation, t.Status, t.UpdatedAt)
	return mapErr(err)
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return scanTask(r.db.Pool.QueryRow(ctx, taskCols+` WHERE id=$1`, id))
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
		hours_awarded=$10,attended_at=$11,completed_at=$12 WHERE id=$1`,
		a.ID, a.Status, a.VolunteerRating, a.VolunteerComment, a.AdminDiscipline, a.AdminExpertise,
		a.AdminEthics, a.AdminComment, a.CompositeScore, a.HoursAwarded, a.AttendedAt, a.CompletedAt)
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
	COALESCE(required_skill_ids, '{}') FROM tasks`

func scanTask(row pgx.Row) (*domain.Task, error) {
	var t domain.Task
	var skills []string
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Location, &t.StartsAt, &t.EndsAt, &t.Capacity, &t.ReservedCount,
		&t.HourWeight, &skills, &t.MinScore, &t.RequiredEducation, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.RequiredSkillIDs)
	if err != nil {
		return nil, mapErr(err)
	}
	t.RequiredSkills = domain.ParseSkillCategories(skills)
	if t.RequiredSkillIDs == nil {
		t.RequiredSkillIDs = []uuid.UUID{}
	}
	return &t, nil
}

const assignmentCols = `SELECT a.id,a.task_id,a.volunteer_id,a.status,a.volunteer_rating,COALESCE(a.volunteer_comment,''),
	a.admin_discipline,a.admin_expertise,a.admin_ethics,COALESCE(a.admin_comment,''),a.composite_score,a.hours_awarded,
	a.attended_at,a.completed_at,a.created_at,
	t.title, t.hour_weight, t.location, t.starts_at,
	v.full_name
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
		&a.Task.Title, &a.Task.HourWeight, &a.Task.Location, &a.Task.StartsAt,
		&a.Volunteer.FullName)
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
