package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type SkillRepo struct{ db *DB }

func (d *DB) Skills() *SkillRepo { return &SkillRepo{d} }

func (r *SkillRepo) ListCatalog(ctx context.Context) ([]domain.SkillGroup, error) {
	grows, err := r.db.Pool.Query(ctx, `SELECT id, slug, title, sort_order FROM skill_groups ORDER BY sort_order, title`)
	if err != nil {
		return nil, err
	}
	defer grows.Close()
	var groups []domain.SkillGroup
	index := map[uuid.UUID]int{}
	for grows.Next() {
		var g domain.SkillGroup
		if err := grows.Scan(&g.ID, &g.Slug, &g.Title, &g.SortOrder); err != nil {
			return nil, err
		}
		g.Skills = []domain.Skill{}
		index[g.ID] = len(groups)
		groups = append(groups, g)
	}
	if err := grows.Err(); err != nil {
		return nil, err
	}
	srows, err := r.db.Pool.Query(ctx, `
		SELECT s.id, s.group_id, s.title, s.status, g.title
		FROM skills s JOIN skill_groups g ON g.id = s.group_id
		ORDER BY g.sort_order, s.title`)
	if err != nil {
		return nil, err
	}
	defer srows.Close()
	for srows.Next() {
		var s domain.Skill
		if err := srows.Scan(&s.ID, &s.GroupID, &s.Title, &s.Status, &s.GroupTitle); err != nil {
			return nil, err
		}
		i, ok := index[s.GroupID]
		if !ok {
			continue
		}
		groups[i].Skills = append(groups[i].Skills, s)
	}
	if groups == nil {
		groups = []domain.SkillGroup{}
	}
	return groups, srows.Err()
}

func (r *SkillRepo) CreateGroup(ctx context.Context, g *domain.SkillGroup) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO skill_groups (id, slug, title, sort_order) VALUES ($1,$2,$3,$4)`,
		g.ID, g.Slug, g.Title, g.SortOrder)
	return mapErr(err)
}

func (r *SkillRepo) UpdateGroup(ctx context.Context, g *domain.SkillGroup) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE skill_groups SET slug=$2, title=$3, sort_order=$4 WHERE id=$1`,
		g.ID, g.Slug, g.Title, g.SortOrder)
	return mapErr(err)
}

func (r *SkillRepo) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE skill_proposals SET created_skill_id=NULL
		WHERE created_skill_id IN (SELECT id FROM skills WHERE group_id=$1)`, id)
	if err != nil {
		return mapErr(err)
	}
	_, err = r.db.Pool.Exec(ctx, `
		UPDATE tasks SET required_skill_ids = COALESCE((
			SELECT ARRAY_AGG(x) FROM unnest(required_skill_ids) AS x
			WHERE x NOT IN (SELECT id FROM skills WHERE group_id=$1)
		), '{}')`, id)
	if err != nil {
		return mapErr(err)
	}
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM skill_groups WHERE id=$1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SkillRepo) CreateSkill(ctx context.Context, s *domain.Skill) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = "active"
	}
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO skills (id, group_id, title, status) VALUES ($1,$2,$3,$4)`,
		s.ID, s.GroupID, s.Title, s.Status)
	return mapErr(err)
}

func (r *SkillRepo) UpdateSkill(ctx context.Context, s *domain.Skill) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE skills SET title=$2, status=$3, group_id=$4 WHERE id=$1`,
		s.ID, s.Title, s.Status, s.GroupID)
	return mapErr(err)
}

func (r *SkillRepo) DeleteSkill(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE skill_proposals SET created_skill_id=NULL WHERE created_skill_id=$1`, id)
	if err != nil {
		return mapErr(err)
	}
	_, err = r.db.Pool.Exec(ctx, `UPDATE tasks SET required_skill_ids = array_remove(required_skill_ids, $1)`, id)
	if err != nil {
		return mapErr(err)
	}
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM skills WHERE id=$1`, id)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SkillRepo) GetSkill(ctx context.Context, id uuid.UUID) (*domain.Skill, error) {
	var s domain.Skill
	err := r.db.Pool.QueryRow(ctx, `
		SELECT s.id, s.group_id, s.title, s.status, g.title
		FROM skills s JOIN skill_groups g ON g.id = s.group_id WHERE s.id=$1`, id).
		Scan(&s.ID, &s.GroupID, &s.Title, &s.Status, &s.GroupTitle)
	if err != nil {
		return nil, mapErr(err)
	}
	return &s, nil
}

func (r *SkillRepo) GetSkillByTitle(ctx context.Context, groupID uuid.UUID, title string) (*domain.Skill, error) {
	var s domain.Skill
	err := r.db.Pool.QueryRow(ctx, `
		SELECT s.id, s.group_id, s.title, s.status, g.title
		FROM skills s JOIN skill_groups g ON g.id = s.group_id
		WHERE s.group_id=$1 AND s.title=$2 LIMIT 1`, groupID, strings.TrimSpace(title)).
		Scan(&s.ID, &s.GroupID, &s.Title, &s.Status, &s.GroupTitle)
	if err != nil {
		return nil, mapErr(err)
	}
	return &s, nil
}

func (r *SkillRepo) GetGroup(ctx context.Context, id uuid.UUID) (*domain.SkillGroup, error) {
	var g domain.SkillGroup
	err := r.db.Pool.QueryRow(ctx, `SELECT id, slug, title, sort_order FROM skill_groups WHERE id=$1`, id).
		Scan(&g.ID, &g.Slug, &g.Title, &g.SortOrder)
	if err != nil {
		return nil, mapErr(err)
	}
	g.Skills = []domain.Skill{}
	return &g, nil
}

func (r *SkillRepo) CreateProposal(ctx context.Context, p *domain.SkillProposal) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Status == "" {
		p.Status = domain.ProposalPending
	}
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO skill_proposals (id, volunteer_id, group_id, title, status, admin_note, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.VolunteerID, p.GroupID, p.Title, p.Status, p.AdminNote, p.CreatedAt)
	return mapErr(err)
}

func (r *SkillRepo) ListProposals(ctx context.Context, status domain.SkillProposalStatus) ([]domain.SkillProposal, error) {
	q := proposalSelect + ` ORDER BY p.created_at DESC`
	args := []any{}
	if status != "" {
		q = proposalSelect + ` WHERE p.status=$1 ORDER BY p.created_at DESC`
		args = append(args, status)
	}
	return r.scanProposals(ctx, q, args...)
}

func (r *SkillRepo) ListProposalsByVolunteer(ctx context.Context, volunteerID uuid.UUID) ([]domain.SkillProposal, error) {
	return r.scanProposals(ctx, proposalSelect+` WHERE p.volunteer_id=$1 ORDER BY p.created_at DESC`, volunteerID)
}

func (r *SkillRepo) GetProposal(ctx context.Context, id uuid.UUID) (*domain.SkillProposal, error) {
	items, err := r.scanProposals(ctx, proposalSelect+` WHERE p.id=$1`, id)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return &items[0], nil
}

func (r *SkillRepo) UpdateProposal(ctx context.Context, p *domain.SkillProposal) error {
	var created any
	if p.CreatedSkillID == uuid.Nil {
		created = nil
	} else {
		created = p.CreatedSkillID
	}
	_, err := r.db.Pool.Exec(ctx, `UPDATE skill_proposals SET group_id=$2, title=$3, status=$4, admin_note=$5, created_skill_id=$6, reviewed_at=now() WHERE id=$1`,
		p.ID, p.GroupID, p.Title, p.Status, p.AdminNote, created)
	return mapErr(err)
}

const proposalSelect = `
	SELECT p.id, p.volunteer_id, v.full_name, p.group_id, g.title, p.title, p.status,
		COALESCE(p.admin_note,''), COALESCE(p.created_skill_id, '00000000-0000-0000-0000-000000000000'::uuid), p.created_at
	FROM skill_proposals p
	JOIN volunteers v ON v.id = p.volunteer_id
	JOIN skill_groups g ON g.id = p.group_id`

func (r *SkillRepo) scanProposals(ctx context.Context, q string, args ...any) ([]domain.SkillProposal, error) {
	rows, err := r.db.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SkillProposal
	for rows.Next() {
		var p domain.SkillProposal
		if err := rows.Scan(&p.ID, &p.VolunteerID, &p.VolunteerName, &p.GroupID, &p.GroupTitle, &p.Title, &p.Status, &p.AdminNote, &p.CreatedSkillID, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []domain.SkillProposal{}
	}
	return out, rows.Err()
}

func (r *SkillRepo) SeedDefaults(ctx context.Context) error {
	type item struct {
		Slug   string
		Title  string
		Skills []string
	}
	catalog := []item{
		{Slug: "general", Title: "عمومی", Skills: []string{"عمومی"}},
		{Slug: "sports", Title: "ورزش", Skills: []string{"فوتبال", "شنا", "والیبال", "بسکتبال", "دو و میدانی"}},
		{Slug: "artistic", Title: "هنر", Skills: []string{"موسیقی", "نقاشی", "تئاتر", "گرافیک", "خوشنویسی"}},
		{Slug: "medical", Title: "پزشکی", Skills: []string{"پزشک", "پرستار", "داروساز", "فوریت‌های پزشکی"}},
		{Slug: "administrative", Title: "اداری", Skills: []string{"منشی", "حسابداری", "بایگانی", "هماهنگی رویداد"}},
		{Slug: "technical", Title: "فنی", Skills: []string{"برنامه‌نویسی", "شبکه", "تعمیرات", "تولید محتوا دیجیتال"}},
		{Slug: "education", Title: "آموزشی", Skills: []string{"معلم", "مربی کودک", "قصه‌گویی", "کمک‌درسی"}},
		{Slug: "logistics", Title: "لجستیک", Skills: []string{"رانندگی", "انبارداری", "حمل و نقل"}},
		{Slug: "psychological", Title: "روان‌شناختی", Skills: []string{"مشاوره", "بازی‌درمانی", "همراهی روانی"}},
		{Slug: "field_ops", Title: "فعالیت‌های جاری", Skills: []string{"بازگشایی قلک", "غرفه‌داری", "جمع‌آوری کمک‌های مردمی", "هماهنگی رویداد میدانی"}},
	}
	for i, g := range catalog {
		var id uuid.UUID
		err := r.db.Pool.QueryRow(ctx, `INSERT INTO skill_groups (id, slug, title, sort_order)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (slug) DO UPDATE SET title=EXCLUDED.title, sort_order=EXCLUDED.sort_order
			RETURNING id`, uuid.New(), g.Slug, g.Title, i).Scan(&id)
		if err != nil {
			return err
		}
		for _, title := range g.Skills {
			_, err := r.db.Pool.Exec(ctx, `INSERT INTO skills (id, group_id, title, status)
				SELECT $1, $2, $3, 'active'
				WHERE NOT EXISTS (SELECT 1 FROM skills s WHERE s.group_id=$2 AND s.title=$3)`,
				uuid.New(), id, title)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
