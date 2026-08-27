package postgres

import (
	"context"
	"encoding/json"
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
	if t.Kind == "" {
		t.Kind = domain.TaskOneOff
	}
	slots, _ := json.Marshal(t.Slots)
	if len(t.Slots) == 0 {
		slots = []byte("[]")
	}
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO tasks (
		id,title,description,location,starts_at,ends_at,capacity,reserved_count,hour_weight,
		required_skills,required_skill_ids,min_score,required_education,work_mode,delivery_hint,status,created_by,created_at,updated_at,
		kind,series_id,weekday,recurrence_slots,requires_training,training_kind,training_location,training_at,training_course_id
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`,
		t.ID, t.Title, t.Description, t.Location, t.StartsAt, t.EndsAt, t.Capacity, t.ReservedCount,
		t.HourWeight, skillsToText(t.RequiredSkills), t.RequiredSkillIDs, t.MinScore, t.RequiredEducation,
		t.WorkMode, t.DeliveryHint, t.Status, nilUUID(t.CreatedBy), t.CreatedAt, t.UpdatedAt,
		t.Kind, nilUUID(t.SeriesID), t.Weekday, slots,
		t.RequiresTraining, t.TrainingKind, t.TrainingLocation, t.TrainingAt, nilUUID(t.TrainingCourseID))
	return mapErr(err)
}

func (r *TaskRepo) Update(ctx context.Context, t *domain.Task) error {
	if t.Kind == "" {
		t.Kind = domain.TaskOneOff
	}
	slots, _ := json.Marshal(t.Slots)
	if len(t.Slots) == 0 {
		slots = []byte("[]")
	}
	_, err := r.db.Pool.Exec(ctx, `UPDATE tasks SET title=$2,description=$3,location=$4,starts_at=$5,ends_at=$6,
		capacity=$7,reserved_count=$8,hour_weight=$9,required_skills=$10,required_skill_ids=$11,min_score=$12,required_education=$13,
		work_mode=$14,delivery_hint=$15,status=$16,updated_at=$17,kind=$18,series_id=$19,weekday=$20,recurrence_slots=$21,
		requires_training=$22,training_kind=$23,training_location=$24,training_at=$25,training_course_id=$26 WHERE id=$1`,
		t.ID, t.Title, t.Description, t.Location, t.StartsAt, t.EndsAt, t.Capacity, t.ReservedCount,
		t.HourWeight, skillsToText(t.RequiredSkills), t.RequiredSkillIDs, t.MinScore, t.RequiredEducation,
		t.WorkMode, t.DeliveryHint, t.Status, t.UpdatedAt, t.Kind, nilUUID(t.SeriesID), t.Weekday, slots,
		t.RequiresTraining, t.TrainingKind, t.TrainingLocation, t.TrainingAt, nilUUID(t.TrainingCourseID))
	return mapErr(err)
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t, err := scanTask(r.db.Pool.QueryRow(ctx, taskCols+` WHERE id=$1`, id))
	if err != nil {
		return nil, err
	}
	r.attachCourses(ctx, []*domain.Task{t})
	return t, nil
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
	if f.Kind != "" {
		where = append(where, fmt.Sprintf("COALESCE(kind,'one_off')=$%d", n))
		args = append(args, f.Kind)
		n++
	}
	if f.ExcludeKind != "" {
		where = append(where, fmt.Sprintf("COALESCE(kind,'one_off') <> $%d", n))
		args = append(args, f.ExcludeKind)
		n++
	}
	if f.SeriesID != uuid.Nil {
		where = append(where, fmt.Sprintf("series_id=$%d", n))
		args = append(args, f.SeriesID)
		n++
	}
	if f.ExcludeVolunteerID != uuid.Nil {
		where = append(where, fmt.Sprintf(`id NOT IN (
			SELECT task_id FROM assignments
			WHERE volunteer_id=$%d AND status IN ('requested','training_pending','reserved','in_progress','attended','submitted','revision_requested','completed')
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
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	ptrs := make([]*domain.Task, len(out))
	for i := range out {
		ptrs[i] = &out[i]
	}
	r.attachCourses(ctx, ptrs)
	return out, total, nil
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

	t, err := scanTask(tx.QueryRow(ctx, taskCols+` WHERE id=$1 FOR UPDATE`, taskID))
	if err != nil {
		return nil, err
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
	a, err := scanAssignment(r.db.Pool.QueryRow(ctx, assignmentCols+` WHERE a.id=$1`, id))
	if err != nil {
		return nil, err
	}
	r.attachAssignmentCourses(ctx, []domain.Assignment{*a})
	if a.Task != nil && a.Task.TrainingCourseID != uuid.Nil {
		if c, err := r.GetTrainingCourse(ctx, a.Task.TrainingCourseID); err == nil {
			a.Task.TrainingCourse = c
		}
	}
	return a, nil
}

func (r *TaskRepo) GetAssignmentByTaskVolunteer(ctx context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	return scanAssignment(r.db.Pool.QueryRow(ctx, assignmentCols+` WHERE a.task_id=$1 AND a.volunteer_id=$2`, taskID, volunteerID))
}

func (r *TaskRepo) UpdateAssignment(ctx context.Context, a *domain.Assignment) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE assignments SET status=$2,volunteer_rating=$3,volunteer_comment=$4,
		admin_discipline=$5,admin_expertise=$6,admin_ethics=$7,admin_comment=$8,composite_score=$9,
		hours_awarded=$10,attended_at=$11,completed_at=$12,delivery_note=$13,delivery_file_name=$14,
		delivery_object_key=$15,delivery_mime=$16,delivered_at=$17,check_in_at=$18,check_out_at=$19 WHERE id=$1`,
		a.ID, a.Status, a.VolunteerRating, a.VolunteerComment, a.AdminDiscipline, a.AdminExpertise,
		a.AdminEthics, a.AdminComment, a.CompositeScore, a.HoursAwarded, a.AttendedAt, a.CompletedAt,
		a.DeliveryNote, a.DeliveryFileName, a.DeliveryObjectKey, a.DeliveryMime, a.DeliveredAt,
		a.CheckInAt, a.CheckOutAt)
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
	if f.SeriesID != uuid.Nil {
		where = append(where, fmt.Sprintf("(t.series_id=$%d OR a.task_id=$%d)", n, n))
		args = append(args, f.SeriesID)
		n++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("a.status=$%d", n))
		args = append(args, f.Status)
		n++
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM assignments a JOIN tasks t ON t.id=a.task_id WHERE "+w, args...).Scan(&total); err != nil {
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
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	r.attachAssignmentCourses(ctx, out)
	return out, total, nil
}

const taskCols = `SELECT id,title,description,COALESCE(location,''),starts_at,ends_at,capacity,reserved_count,hour_weight,
	required_skills,min_score,COALESCE(required_education,''),status,COALESCE(created_by,'00000000-0000-0000-0000-000000000000'),created_at,updated_at,
	COALESCE(required_skill_ids, '{}'), COALESCE(work_mode,'onsite'), COALESCE(delivery_hint,''),
	COALESCE(kind,'one_off'), COALESCE(series_id, '00000000-0000-0000-0000-000000000000'), COALESCE(weekday, 0), COALESCE(recurrence_slots, '[]'),
	COALESCE(requires_training,false), COALESCE(training_kind,''), COALESCE(training_location,''), training_at,
	COALESCE(training_course_id, '00000000-0000-0000-0000-000000000000') FROM tasks`

func scanTask(row pgx.Row) (*domain.Task, error) {
	var t domain.Task
	var skills []string
	var slots []byte
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Location, &t.StartsAt, &t.EndsAt, &t.Capacity, &t.ReservedCount,
		&t.HourWeight, &skills, &t.MinScore, &t.RequiredEducation, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.RequiredSkillIDs,
		&t.WorkMode, &t.DeliveryHint, &t.Kind, &t.SeriesID, &t.Weekday, &slots,
		&t.RequiresTraining, &t.TrainingKind, &t.TrainingLocation, &t.TrainingAt, &t.TrainingCourseID)
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
	if t.Kind == "" {
		t.Kind = domain.TaskOneOff
	}
	if len(slots) > 0 && string(slots) != "[]" && string(slots) != "null" {
		_ = json.Unmarshal(slots, &t.Slots)
	}
	if t.Slots == nil {
		t.Slots = []domain.TaskSlot{}
	}
	return &t, nil
}

const assignmentCols = `SELECT a.id,a.task_id,a.volunteer_id,a.status,a.volunteer_rating,COALESCE(a.volunteer_comment,''),
	a.admin_discipline,a.admin_expertise,a.admin_ethics,COALESCE(a.admin_comment,''),a.composite_score,a.hours_awarded,
	a.attended_at,a.check_in_at,a.check_out_at,a.completed_at,a.created_at,
	COALESCE(a.delivery_note,''), COALESCE(a.delivery_file_name,''), COALESCE(a.delivery_object_key,''), COALESCE(a.delivery_mime,''), a.delivered_at,
	t.title, COALESCE(t.description,''), t.hour_weight, COALESCE(t.location,''), t.starts_at, t.ends_at, COALESCE(t.work_mode,'onsite'), COALESCE(t.delivery_hint,''),
	COALESCE(t.kind,'one_off'), COALESCE(t.series_id, '00000000-0000-0000-0000-000000000000'), COALESCE(t.weekday, 0),
	COALESCE(t.requires_training,false), COALESCE(t.training_kind,''), COALESCE(t.training_location,''), t.training_at,
	COALESCE(t.training_course_id, '00000000-0000-0000-0000-000000000000'),
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
		&a.AttendedAt, &a.CheckInAt, &a.CheckOutAt, &a.CompletedAt, &a.CreatedAt,
		&a.DeliveryNote, &a.DeliveryFileName, &a.DeliveryObjectKey, &a.DeliveryMime, &a.DeliveredAt,
		&a.Task.Title, &a.Task.Description, &a.Task.HourWeight, &a.Task.Location, &a.Task.StartsAt, &a.Task.EndsAt, &a.Task.WorkMode, &a.Task.DeliveryHint,
		&a.Task.Kind, &a.Task.SeriesID, &a.Task.Weekday,
		&a.Task.RequiresTraining, &a.Task.TrainingKind, &a.Task.TrainingLocation, &a.Task.TrainingAt, &a.Task.TrainingCourseID,
		&a.Volunteer.FullName, &a.Volunteer.Phone)
	if err != nil {
		return nil, mapErr(err)
	}
	a.Task.ID = a.TaskID
	a.Volunteer.ID = a.VolunteerID
	return &a, nil
}

func (r *TaskRepo) CreateVolunteerTraining(ctx context.Context, t *domain.VolunteerTraining) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO volunteer_trainings (
		id, volunteer_id, series_id, training_kind, training_location, training_at,
		source_task_id, assignment_id, confirmed_by, confirmed_at, course_id, course_title
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		t.ID, t.VolunteerID, nilUUID(t.SeriesID), t.TrainingKind, t.TrainingLocation, t.TrainingAt,
		nilUUID(t.SourceTaskID), nilUUID(t.AssignmentID), nilUUID(t.ConfirmedBy), t.ConfirmedAt,
		nilUUID(t.CourseID), t.CourseTitle)
	return mapErr(err)
}

func (r *TaskRepo) ListVolunteerTrainings(ctx context.Context, volunteerID uuid.UUID) ([]domain.VolunteerTraining, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT vt.id, vt.volunteer_id, COALESCE(vt.series_id, '00000000-0000-0000-0000-000000000000'),
		COALESCE(vt.training_kind,''), COALESCE(vt.training_location,''), vt.training_at,
		COALESCE(vt.source_task_id, '00000000-0000-0000-0000-000000000000'), COALESCE(t.title,''),
		COALESCE(vt.assignment_id, '00000000-0000-0000-0000-000000000000'),
		COALESCE(vt.confirmed_by, '00000000-0000-0000-0000-000000000000'), vt.confirmed_at,
		COALESCE(vt.course_id, '00000000-0000-0000-0000-000000000000'),
		COALESCE(NULLIF(vt.course_title,''), COALESCE(c.title,''), COALESCE(t.title,''))
		FROM volunteer_trainings vt
		LEFT JOIN tasks t ON t.id = vt.source_task_id
		LEFT JOIN training_courses c ON c.id = vt.course_id
		WHERE vt.volunteer_id=$1
		ORDER BY vt.confirmed_at DESC`, volunteerID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []domain.VolunteerTraining
	for rows.Next() {
		var x domain.VolunteerTraining
		if err := rows.Scan(&x.ID, &x.VolunteerID, &x.SeriesID, &x.TrainingKind, &x.TrainingLocation, &x.TrainingAt,
			&x.SourceTaskID, &x.SourceTaskTitle, &x.AssignmentID, &x.ConfirmedBy, &x.ConfirmedAt,
			&x.CourseID, &x.CourseTitle); err != nil {
			return nil, mapErr(err)
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *TaskRepo) HasCompletedTraining(ctx context.Context, volunteerID uuid.UUID, t *domain.Task) (bool, error) {
	if t == nil {
		return false, nil
	}
	sid := t.TrainingSeriesID()
	title := t.TrainingCourseTitle()
	var n int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM volunteer_trainings
		WHERE volunteer_id=$1 AND (
			($5::uuid IS NOT NULL AND course_id = $5)
			OR (
				btrim($6) <> '' AND lower(btrim(course_title)) = lower(btrim($6))
			)
			OR ($2::uuid IS NOT NULL AND series_id = $2)
			OR (
				lower(btrim(training_kind)) = lower(btrim($3))
				AND lower(btrim(training_location)) = lower(btrim($4))
				AND btrim($3) <> '' AND btrim($4) <> ''
			)
		)`, volunteerID, nilUUID(sid), t.TrainingKind, t.TrainingLocation, nilUUID(t.TrainingCourseID), title).Scan(&n)
	if err != nil {
		return false, mapErr(err)
	}
	return n > 0, nil
}

func (r *TaskRepo) AddAssignmentEvent(ctx context.Context, e *domain.AssignmentEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO assignment_events (id, assignment_id, kind, note, actor_role, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`, e.ID, e.AssignmentID, e.Kind, e.Note, e.ActorRole, e.CreatedAt); err != nil {
		return mapErr(err)
	}
	for i := range e.Files {
		f := &e.Files[i]
		if f.ID == uuid.Nil {
			f.ID = uuid.New()
		}
		f.EventID = e.ID
		if _, err := tx.Exec(ctx, `INSERT INTO assignment_event_files (id, event_id, file_name, object_key, mime_type, size_bytes)
			VALUES ($1,$2,$3,$4,$5,$6)`, f.ID, e.ID, f.FileName, f.ObjectKey, f.MimeType, f.SizeBytes); err != nil {
			return mapErr(err)
		}
	}
	return tx.Commit(ctx)
}

func (r *TaskRepo) ListAssignmentEvents(ctx context.Context, assignmentIDs []uuid.UUID) (map[uuid.UUID][]domain.AssignmentEvent, error) {
	out := map[uuid.UUID][]domain.AssignmentEvent{}
	if len(assignmentIDs) == 0 {
		return out, nil
	}
	rows, err := r.db.Pool.Query(ctx, `SELECT id, assignment_id, kind, COALESCE(note,''), COALESCE(actor_role,''), created_at
		FROM assignment_events WHERE assignment_id = ANY($1) ORDER BY created_at ASC, id ASC`, assignmentIDs)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var eventIDs []uuid.UUID
	for rows.Next() {
		var e domain.AssignmentEvent
		if err := rows.Scan(&e.ID, &e.AssignmentID, &e.Kind, &e.Note, &e.ActorRole, &e.CreatedAt); err != nil {
			return nil, mapErr(err)
		}
		e.Files = []domain.AssignmentEventFile{}
		eventIDs = append(eventIDs, e.ID)
		out[e.AssignmentID] = append(out[e.AssignmentID], e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(eventIDs) == 0 {
		return out, nil
	}
	frows, err := r.db.Pool.Query(ctx, `SELECT id, event_id, COALESCE(file_name,''), COALESCE(object_key,''), COALESCE(mime_type,''), size_bytes
		FROM assignment_event_files WHERE event_id = ANY($1) ORDER BY file_name`, eventIDs)
	if err != nil {
		return nil, mapErr(err)
	}
	defer frows.Close()
	filesByEvent := map[uuid.UUID][]domain.AssignmentEventFile{}
	for frows.Next() {
		var f domain.AssignmentEventFile
		if err := frows.Scan(&f.ID, &f.EventID, &f.FileName, &f.ObjectKey, &f.MimeType, &f.SizeBytes); err != nil {
			return nil, mapErr(err)
		}
		filesByEvent[f.EventID] = append(filesByEvent[f.EventID], f)
	}
	if err := frows.Err(); err != nil {
		return nil, err
	}
	for asgID, events := range out {
		for i := range events {
			events[i].Files = filesByEvent[events[i].ID]
			if events[i].Files == nil {
				events[i].Files = []domain.AssignmentEventFile{}
			}
		}
		out[asgID] = events
	}
	return out, nil
}

func (r *TaskRepo) GetAssignmentFile(ctx context.Context, fileID uuid.UUID) (*domain.AssignmentEventFile, error) {
	var f domain.AssignmentEventFile
	err := r.db.Pool.QueryRow(ctx, `SELECT f.id, f.event_id, e.assignment_id, COALESCE(f.file_name,''), COALESCE(f.object_key,''), COALESCE(f.mime_type,''), f.size_bytes
		FROM assignment_event_files f
		JOIN assignment_events e ON e.id = f.event_id
		WHERE f.id=$1`, fileID).Scan(&f.ID, &f.EventID, &f.AssignmentID, &f.FileName, &f.ObjectKey, &f.MimeType, &f.SizeBytes)
	if err != nil {
		return nil, mapErr(err)
	}
	return &f, nil
}

func nilUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
