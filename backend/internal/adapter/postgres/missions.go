package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type MissionRepo struct{ db *DB }

func (d *DB) Missions() *MissionRepo { return &MissionRepo{d} }

func (r *MissionRepo) Create(ctx context.Context, m *domain.Mission) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO missions (id,title,description,kind,hour_weight,deadline_hours,webhook_event,target_count,verify_mode,verify_url,verify_token,status,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		m.ID, m.Title, m.Description, m.Kind, m.HourWeight, m.DeadlineHours, m.WebhookEvent, m.TargetCount, m.VerifyMode, m.VerifyURL, m.VerifyToken, m.Status, m.CreatedAt)
	return mapErr(err)
}

func (r *MissionRepo) Update(ctx context.Context, m *domain.Mission) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE missions SET title=$2,description=$3,kind=$4,hour_weight=$5,deadline_hours=$6,webhook_event=$7,target_count=$8,verify_mode=$9,verify_url=$10,verify_token=$11,status=$12 WHERE id=$1`,
		m.ID, m.Title, m.Description, m.Kind, m.HourWeight, m.DeadlineHours, m.WebhookEvent, m.TargetCount, m.VerifyMode, m.VerifyURL, m.VerifyToken, m.Status)
	return mapErr(err)
}

func (r *MissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Mission, error) {
	return scanMission(r.db.Pool.QueryRow(ctx, missionCols+` WHERE id=$1`, id))
}

func (r *MissionRepo) List(ctx context.Context, activeOnly bool) ([]domain.Mission, error) {
	q := missionCols + ` ORDER BY created_at DESC`
	if activeOnly {
		q = missionCols + ` WHERE status='active' ORDER BY created_at DESC`
	}
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Mission
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *MissionRepo) GetByWebhookEvent(ctx context.Context, event string) ([]domain.Mission, error) {
	rows, err := r.db.Pool.Query(ctx, missionCols+` WHERE webhook_event=$1 AND status='active'`, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Mission
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *MissionRepo) GetByVerifyToken(ctx context.Context, token string) (*domain.Mission, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, domain.ErrNotFound
	}
	return scanMission(r.db.Pool.QueryRow(ctx, missionCols+` WHERE verify_token=$1 AND status='active'`, token))
}

func (r *MissionRepo) UpsertProgress(ctx context.Context, p *domain.MissionProgress) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO mission_progress (id,mission_id,volunteer_id,status,progress,started_at,due_at,completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (mission_id, volunteer_id) DO UPDATE SET status=EXCLUDED.status, progress=EXCLUDED.progress, due_at=EXCLUDED.due_at, completed_at=EXCLUDED.completed_at`,
		p.ID, p.MissionID, p.VolunteerID, p.Status, p.Progress, p.StartedAt, p.DueAt, p.CompletedAt)
	return mapErr(err)
}

func (r *MissionRepo) GetProgress(ctx context.Context, missionID, volunteerID uuid.UUID) (*domain.MissionProgress, error) {
	return scanProgress(r.db.Pool.QueryRow(ctx, progressCols+` WHERE p.mission_id=$1 AND p.volunteer_id=$2`, missionID, volunteerID))
}

func (r *MissionRepo) ListProgressByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]domain.MissionProgress, error) {
	rows, err := r.db.Pool.Query(ctx, progressCols+` WHERE p.volunteer_id=$1 ORDER BY p.started_at DESC`, volunteerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MissionProgress
	for rows.Next() {
		p, err := scanProgress(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

const missionCols = `SELECT id,title,description,kind,hour_weight,deadline_hours,COALESCE(webhook_event,''),target_count,COALESCE(verify_mode,'internal'),COALESCE(verify_url,''),COALESCE(verify_token,''),status,created_at FROM missions`

func scanMission(row pgx.Row) (*domain.Mission, error) {
	var m domain.Mission
	err := row.Scan(&m.ID, &m.Title, &m.Description, &m.Kind, &m.HourWeight, &m.DeadlineHours, &m.WebhookEvent, &m.TargetCount, &m.VerifyMode, &m.VerifyURL, &m.VerifyToken, &m.Status, &m.CreatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	if m.VerifyMode == "" {
		m.VerifyMode = domain.VerifyInternal
	}
	return &m, nil
}

const progressCols = `SELECT p.id,p.mission_id,p.volunteer_id,p.status,p.progress,p.started_at,p.due_at,p.completed_at,
	m.title,COALESCE(m.description,''),m.kind,m.hour_weight,m.target_count,COALESCE(m.verify_mode,'internal'),m.deadline_hours
	FROM mission_progress p JOIN missions m ON m.id=p.mission_id`

func scanProgress(row pgx.Row) (*domain.MissionProgress, error) {
	var p domain.MissionProgress
	p.Mission = &domain.Mission{}
	err := row.Scan(&p.ID, &p.MissionID, &p.VolunteerID, &p.Status, &p.Progress, &p.StartedAt, &p.DueAt, &p.CompletedAt,
		&p.Mission.Title, &p.Mission.Description, &p.Mission.Kind, &p.Mission.HourWeight, &p.Mission.TargetCount, &p.Mission.VerifyMode, &p.Mission.DeadlineHours)
	if err != nil {
		return nil, mapErr(err)
	}
	p.Mission.ID = p.MissionID
	if p.Mission.VerifyMode == "" {
		p.Mission.VerifyMode = domain.VerifyInternal
	}
	return &p, nil
}

type CertRepo struct{ db *DB }

func (d *DB) Certificates() *CertRepo { return &CertRepo{d} }

func (r *CertRepo) Create(ctx context.Context, c *domain.Certificate) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO certificates (id,verification_code,volunteer_id,kind,assignment_id,title,hours,period_start,period_end,issued_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.VerificationCode, c.VolunteerID, c.Kind, c.AssignmentID, c.Title, c.Hours, c.PeriodStart, c.PeriodEnd, c.IssuedAt)
	return mapErr(err)
}

func (r *CertRepo) GetByVerificationCode(ctx context.Context, code uuid.UUID) (*domain.Certificate, error) {
	return scanCert(r.db.Pool.QueryRow(ctx, certCols+` WHERE verification_code=$1`, code))
}

func (r *CertRepo) GetByAssignment(ctx context.Context, assignmentID uuid.UUID) (*domain.Certificate, error) {
	return scanCert(r.db.Pool.QueryRow(ctx, certCols+` WHERE assignment_id=$1 ORDER BY issued_at DESC LIMIT 1`, assignmentID))
}

func (r *CertRepo) ListByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]domain.Certificate, error) {
	rows, err := r.db.Pool.Query(ctx, certCols+` WHERE volunteer_id=$1 ORDER BY issued_at DESC`, volunteerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		c, err := scanCert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *CertRepo) ExistsForAssignment(ctx context.Context, assignmentID uuid.UUID) (bool, error) {
	var n int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM certificates WHERE assignment_id=$1`, assignmentID).Scan(&n)
	return n > 0, err
}

const certCols = `SELECT id,verification_code,volunteer_id,kind,assignment_id,title,hours,period_start,period_end,issued_at FROM certificates`

func scanCert(row pgx.Row) (*domain.Certificate, error) {
	var c domain.Certificate
	err := row.Scan(&c.ID, &c.VerificationCode, &c.VolunteerID, &c.Kind, &c.AssignmentID, &c.Title, &c.Hours, &c.PeriodStart, &c.PeriodEnd, &c.IssuedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &c, nil
}

type NotifyRepo struct{ db *DB }

func (d *DB) Notifications() *NotifyRepo { return &NotifyRepo{d} }

func (r *NotifyRepo) Create(ctx context.Context, n *domain.Notification) error {
	if n.Kind == "" {
		n.Kind = domain.NotifyNotice
	}
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO notifications (id,user_id,title,body,read,created_at,kind,remind_at,fired_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		n.ID, n.UserID, n.Title, n.Body, n.Read, n.CreatedAt, n.Kind, n.RemindAt, n.FiredAt)
	return mapErr(err)
}

func (r *NotifyRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	_ = r.FireDueReminders(ctx, time.Now().UTC())
	rows, err := r.db.Pool.Query(ctx, `SELECT id,user_id,title,body,read,created_at,COALESCE(kind,'notice'),remind_at,fired_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Read, &n.CreatedAt, &n.Kind, &n.RemindAt, &n.FiredAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *NotifyRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE notifications SET read=true WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (r *NotifyRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE notifications SET read=true WHERE user_id=$1 AND read=false`, userID)
	return err
}

func (r *NotifyRepo) NotifyStaff(ctx context.Context, title, body string) error {
	rows, err := r.db.Pool.Query(ctx, `SELECT id FROM users WHERE role IN ('admin','operator')`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		_ = r.Notify(ctx, id, title, body)
	}
	return nil
}

func (r *NotifyRepo) Notify(ctx context.Context, userID uuid.UUID, title, body string) error {
	n := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Body:      body,
		Kind:      domain.NotifyNotice,
		CreatedAt: time.Now().UTC(),
	}
	return r.Create(ctx, n)
}

func (r *NotifyRepo) NotifyReminder(ctx context.Context, userID uuid.UUID, title, body string, remindAt time.Time) error {
	now := time.Now().UTC()
	n := &domain.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Body:      body,
		Kind:      domain.NotifyReminder,
		RemindAt:  &remindAt,
		Read:      true,
		CreatedAt: now,
	}
	if !remindAt.After(now) {
		n.Read = false
		n.FiredAt = &now
		n.Title = "زمان آموزش فرا رسیده"
	}
	return r.Create(ctx, n)
}

func (r *NotifyRepo) FireDueReminders(ctx context.Context, now time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE notifications
		SET read=false, fired_at=$1, title='زمان آموزش فرا رسیده'
		WHERE kind='reminder' AND remind_at IS NOT NULL AND remind_at <= $1 AND fired_at IS NULL`, now)
	return err
}

type StatsRepo struct{ db *DB }

func (d *DB) Stats() *StatsRepo { return &StatsRepo{d} }

func (r *StatsRepo) Dashboard(ctx context.Context) (*domain.DashboardStats, error) {
	s := &domain.DashboardStats{SkillDistribution: map[string]int{}}
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM volunteers`).Scan(&s.TotalVolunteers)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM volunteers WHERE status='pending'`).Scan(&s.PendingVolunteers)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM volunteers WHERE status='approved'`).Scan(&s.ApprovedVolunteers)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM tasks WHERE status='open' AND ends_at > now()`).Scan(&s.OpenTasks)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM assignments WHERE status IN ('training_pending','reserved','in_progress','attended')`).Scan(&s.ActiveAssignments)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM assignments WHERE status='completed' AND completed_at >= date_trunc('month', now())`).Scan(&s.CompletedThisMonth)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_hours),0) FROM volunteers`).Scan(&s.TotalHours)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM assignments a JOIN tasks t ON t.id=a.task_id WHERE a.status='requested'`).Scan(&s.PendingTaskRequests)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM assignments a JOIN tasks t ON t.id=a.task_id WHERE a.status='training_pending'`).Scan(&s.PendingTrainingConfirmations)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM assignments a JOIN tasks t ON t.id=a.task_id WHERE a.status='submitted'`).Scan(&s.PendingDeliveries)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM skill_proposals p JOIN volunteers v ON v.id=p.volunteer_id JOIN skill_groups g ON g.id=p.group_id WHERE p.status='pending'`).Scan(&s.PendingSkillProposals)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM certificate_requests r JOIN volunteers v ON v.id=r.volunteer_id WHERE r.status IN ('pending','preparing','ready')`).Scan(&s.PendingCertificates)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM tickets t JOIN volunteers v ON v.id=t.volunteer_id WHERE t.status='open'`).Scan(&s.OpenTickets)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM volunteers v WHERE `+resubmittedDocsSQL).Scan(&s.ResubmittedDocuments)
	if s.ApprovedVolunteers > 0 {
		s.ParticipationRate = float64(s.ActiveAssignments) / float64(s.ApprovedVolunteers)
	}
	s.OnlineEstimate = s.ActiveAssignments
	s.SkillDistribution = r.skillCounts(ctx)
	return s, nil
}

func (r *StatsRepo) skillCounts(ctx context.Context) map[string]int {
	out := map[string]int{}
	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			COALESCE(NULLIF(sg.title, ''), NULLIF(sk.title, ''), skill) AS label,
			COUNT(*)
		FROM volunteers v
		CROSS JOIN LATERAL unnest(v.skill_categories) AS skill
		LEFT JOIN skill_groups sg ON sg.slug = skill
		LEFT JOIN skills sk ON sk.id::text = skill
		WHERE v.status = 'approved' AND COALESCE(skill, '') <> ''
		GROUP BY 1`)
	if err != nil {
		rows, err = r.db.Pool.Query(ctx, `SELECT unnest(skill_categories) skill, COUNT(*) FROM volunteers WHERE status='approved' GROUP BY skill`)
		if err != nil {
			return out
		}
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var c int
		if err := rows.Scan(&k, &c); err != nil {
			continue
		}
		out[k] += c
	}
	return out
}

func (r *StatsRepo) Ranking(ctx context.Context, limit int) ([]domain.RankingRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.Pool.Query(ctx, `SELECT id,full_name,COALESCE(city,''),skill_categories,average_score,total_hours,completed_tasks,status
		FROM volunteers WHERE status IN ('approved','suspended') ORDER BY total_hours DESC, average_score DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RankingRow
	for rows.Next() {
		var row domain.RankingRow
		var skills []string
		if err := rows.Scan(&row.VolunteerID, &row.FullName, &row.City, &skills, &row.AverageScore, &row.TotalHours, &row.CompletedTasks, &row.Status); err != nil {
			return nil, err
		}
		row.Skills = domain.ParseSkillCategories(skills)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *StatsRepo) SkillDistribution(ctx context.Context) (map[string]int, error) {
	s, err := r.Dashboard(ctx)
	if err != nil {
		return nil, err
	}
	return s.SkillDistribution, nil
}

func countBy(ctx context.Context, r *StatsRepo, q string) map[string]int {
	out := map[string]int{}
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var n int
		if rows.Scan(&k, &n) == nil {
			out[k] = n
		}
	}
	return out
}

func (r *StatsRepo) Overview(ctx context.Context) (*domain.ReportOverview, error) {
	dash, err := r.Dashboard(ctx)
	if err != nil {
		return nil, err
	}
	o := &domain.ReportOverview{DashboardStats: *dash}
	o.VolunteersByStatus = countBy(ctx, r, `SELECT status, COUNT(*) FROM volunteers GROUP BY status`)
	o.AssignmentsByStatus = countBy(ctx, r, `SELECT status, COUNT(*) FROM assignments GROUP BY status`)
	o.TasksByStatus = countBy(ctx, r, `SELECT status, COUNT(*) FROM tasks WHERE COALESCE(kind,'one_off') <> 'occurrence' GROUP BY status`)
	o.TasksByKind = countBy(ctx, r, `SELECT COALESCE(kind,'one_off'), COUNT(*) FROM tasks GROUP BY 1`)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COALESCE(SUM(hours_awarded),0) FROM assignments WHERE status='completed' AND completed_at >= date_trunc('month', now())`).Scan(&o.HoursThisMonth)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM certificates`).Scan(&o.CertificatesIssued)
	rows, err := r.db.Pool.Query(ctx, `SELECT COALESCE(NULLIF(city,''),'نامشخص') city, COUNT(*) FROM volunteers WHERE status='approved' GROUP BY 1 ORDER BY COUNT(*) DESC LIMIT 8`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c domain.CityCount
			if rows.Scan(&c.City, &c.Count) == nil {
				o.TopCities = append(o.TopCities, c)
			}
		}
	}
	if o.TopCities == nil {
		o.TopCities = []domain.CityCount{}
	}
	return o, nil
}
