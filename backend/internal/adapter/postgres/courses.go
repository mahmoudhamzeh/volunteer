package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

const courseCols = `SELECT id, title, COALESCE(kind,'in_person'), COALESCE(location,''), training_at,
	COALESCE(description,''), COALESCE(status,'active'), created_at, updated_at FROM training_courses`

func scanCourse(row pgx.Row) (*domain.TrainingCourse, error) {
	var c domain.TrainingCourse
	err := row.Scan(&c.ID, &c.Title, &c.Kind, &c.Location, &c.TrainingAt, &c.Description, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	if c.Kind == "" {
		c.Kind = domain.TrainingInPerson
	}
	if c.Status == "" {
		c.Status = domain.TrainingCourseActive
	}
	return &c, nil
}

func (r *TaskRepo) CreateTrainingCourse(ctx context.Context, c *domain.TrainingCourse) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO training_courses (id,title,kind,location,training_at,description,status,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, c.Title, c.Kind, c.Location, c.TrainingAt, c.Description, c.Status, c.CreatedAt, c.UpdatedAt)
	return mapErr(err)
}

func (r *TaskRepo) UpdateTrainingCourse(ctx context.Context, c *domain.TrainingCourse) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE training_courses SET title=$2,kind=$3,location=$4,training_at=$5,description=$6,status=$7,updated_at=$8 WHERE id=$1`,
		c.ID, c.Title, c.Kind, c.Location, c.TrainingAt, c.Description, c.Status, c.UpdatedAt)
	return mapErr(err)
}

func (r *TaskRepo) GetTrainingCourse(ctx context.Context, id uuid.UUID) (*domain.TrainingCourse, error) {
	if id == uuid.Nil {
		return nil, domain.ErrNotFound
	}
	return scanCourse(r.db.Pool.QueryRow(ctx, courseCols+` WHERE id=$1`, id))
}

func (r *TaskRepo) ListTrainingCourses(ctx context.Context, activeOnly bool) ([]domain.TrainingCourse, error) {
	q := courseCols + ` ORDER BY title`
	if activeOnly {
		q = courseCols + ` WHERE status='active' ORDER BY title`
	}
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []domain.TrainingCourse
	for rows.Next() {
		c, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	if out == nil {
		out = []domain.TrainingCourse{}
	}
	return out, rows.Err()
}

func (r *TaskRepo) attachCourses(ctx context.Context, items []*domain.Task) {
	ids := courseIDsFromTasks(items)
	if len(ids) == 0 {
		return
	}
	by := r.loadCourses(ctx, ids)
	for _, t := range items {
		if t == nil || t.TrainingCourseID == uuid.Nil {
			continue
		}
		if c, ok := by[t.TrainingCourseID]; ok {
			cp := c
			t.TrainingCourse = &cp
		}
	}
}

func (r *TaskRepo) attachAssignmentCourses(ctx context.Context, items []domain.Assignment) {
	ptrs := make([]*domain.Task, 0, len(items))
	for i := range items {
		if items[i].Task != nil {
			ptrs = append(ptrs, items[i].Task)
		}
	}
	r.attachCourses(ctx, ptrs)
}

func courseIDsFromTasks(items []*domain.Task) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	var ids []uuid.UUID
	for _, t := range items {
		if t == nil || t.TrainingCourseID == uuid.Nil {
			continue
		}
		if _, ok := seen[t.TrainingCourseID]; ok {
			continue
		}
		seen[t.TrainingCourseID] = struct{}{}
		ids = append(ids, t.TrainingCourseID)
	}
	return ids
}

func (r *TaskRepo) loadCourses(ctx context.Context, ids []uuid.UUID) map[uuid.UUID]domain.TrainingCourse {
	out := map[uuid.UUID]domain.TrainingCourse{}
	if len(ids) == 0 {
		return out
	}
	rows, err := r.db.Pool.Query(ctx, courseCols+` WHERE id = ANY($1)`, ids)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		c, err := scanCourse(rows)
		if err != nil {
			continue
		}
		out[c.ID] = *c
	}
	return out
}
