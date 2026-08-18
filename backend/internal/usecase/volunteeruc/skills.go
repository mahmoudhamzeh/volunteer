package volunteeruc

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func (s *Service) Catalog(ctx context.Context) ([]domain.SkillGroup, error) {
	if s.skills == nil {
		return []domain.SkillGroup{}, nil
	}
	return s.skills.ListCatalog(ctx)
}

func (s *Service) ProposeSkill(ctx context.Context, userID, groupID uuid.UUID, title string) (*domain.SkillProposal, error) {
	if s.skills == nil {
		return nil, domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, domain.Invalid("عنوان مهارت الزامی است")
	}
	if groupID == uuid.Nil {
		return nil, domain.Invalid("گروه مهارت را انتخاب کنید")
	}
	g, err := s.skills.GetGroup(ctx, groupID)
	if err != nil {
		return nil, domain.Invalid("گروه مهارت نامعتبر است")
	}
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing, err := s.skills.GetSkillByTitle(ctx, groupID, title); err == nil && existing != nil {
		return nil, domain.Invalid("این مهارت در فهرست وجود دارد؛ آن را از لیست انتخاب کنید")
	} else if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	p := &domain.SkillProposal{
		ID:          uuid.New(),
		VolunteerID: v.ID,
		GroupID:     g.ID,
		GroupTitle:  g.Title,
		Title:       title,
		Status:      domain.ProposalPending,
		CreatedAt:   s.clock.Now(),
	}
	if err := s.skills.CreateProposal(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) MyProposals(ctx context.Context, userID uuid.UUID) ([]domain.SkillProposal, error) {
	if s.skills == nil {
		return []domain.SkillProposal{}, nil
	}
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.skills.ListProposalsByVolunteer(ctx, v.ID)
}

func (s *Service) ListProposals(ctx context.Context, status domain.SkillProposalStatus) ([]domain.SkillProposal, error) {
	if s.skills == nil {
		return []domain.SkillProposal{}, nil
	}
	return s.skills.ListProposals(ctx, status)
}

type ProposalReviewInput struct {
	Action    string `json:"action"`
	Title     string `json:"title"`
	GroupID   string `json:"group_id"`
	AdminNote string `json:"admin_note"`
}

func (s *Service) ReviewProposal(ctx context.Context, proposalID uuid.UUID, in ProposalReviewInput) (*domain.SkillProposal, error) {
	if s.skills == nil {
		return nil, domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	p, err := s.skills.GetProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.ProposalPending {
		return nil, domain.Invalid("این پیشنهاد قبلا بررسی شده است")
	}
	switch in.Action {
	case "reject":
		p.Status = domain.ProposalRejected
		p.AdminNote = strings.TrimSpace(in.AdminNote)
		if p.AdminNote == "" {
			return nil, domain.Invalid("دلیل رد را وارد کنید")
		}
		if err := s.skills.UpdateProposal(ctx, p); err != nil {
			return nil, err
		}
		return p, nil
	case "approve", "edit_approve":
		title := strings.TrimSpace(in.Title)
		if title == "" {
			title = p.Title
		}
		groupID := p.GroupID
		if in.GroupID != "" {
			gid, err := uuid.Parse(in.GroupID)
			if err != nil {
				return nil, domain.Invalid("گروه مهارت نامعتبر است")
			}
			if _, err := s.skills.GetGroup(ctx, gid); err != nil {
				return nil, domain.Invalid("گروه مهارت نامعتبر است")
			}
			groupID = gid
		}
		p.Title = title
		p.GroupID = groupID
		p.AdminNote = strings.TrimSpace(in.AdminNote)
		sk, err := s.skills.GetSkillByTitle(ctx, groupID, title)
		if err == domain.ErrNotFound {
			g, gerr := s.skills.GetGroup(ctx, groupID)
			if gerr != nil {
				return nil, gerr
			}
			sk = &domain.Skill{ID: uuid.New(), GroupID: groupID, Title: title, Status: "active", GroupTitle: g.Title}
			if err := s.skills.CreateSkill(ctx, sk); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		existing, err := s.volunteers.ListVolunteerSkills(ctx, p.VolunteerID)
		if err != nil {
			return nil, err
		}
		ids := make([]uuid.UUID, 0, len(existing)+1)
		seen := map[uuid.UUID]struct{}{}
		for _, e := range existing {
			ids = append(ids, e.SkillID)
			seen[e.SkillID] = struct{}{}
		}
		if _, ok := seen[sk.ID]; !ok {
			ids = append(ids, sk.ID)
		}
		if err := s.volunteers.ReplaceSkills(ctx, p.VolunteerID, ids); err != nil {
			return nil, err
		}
		vol, err := s.volunteers.GetByID(ctx, p.VolunteerID)
		if err != nil {
			return nil, err
		}
		if err := s.syncCategories(ctx, vol); err != nil {
			return nil, err
		}
		vol.UpdatedAt = s.clock.Now()
		if err := s.volunteers.Update(ctx, vol); err != nil {
			return nil, err
		}
		p.Status = domain.ProposalApproved
		p.CreatedSkillID = sk.ID
		if g, err := s.skills.GetGroup(ctx, groupID); err == nil {
			p.GroupTitle = g.Title
		}
		if err := s.skills.UpdateProposal(ctx, p); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, domain.Invalid("عملیات نامعتبر است")
	}
}

func (s *Service) CreateGroup(ctx context.Context, slug, title string, sortOrder int) (*domain.SkillGroup, error) {
	if s.skills == nil {
		return nil, domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, domain.Invalid("عنوان گروه الزامی است")
	}
	slug = groupSlug(slug, title)
	g := &domain.SkillGroup{ID: uuid.New(), Slug: slug, Title: title, SortOrder: sortOrder, Skills: []domain.Skill{}}
	if err := s.skills.CreateGroup(ctx, g); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			g.Slug = groupSlug("", title+"-"+uuid.NewString()[:6])
			g.ID = uuid.New()
			if err := s.skills.CreateGroup(ctx, g); err != nil {
				return nil, err
			}
			return g, nil
		}
		return nil, err
	}
	return g, nil
}

func groupSlug(slug, title string) string {
	s := strings.ToLower(strings.TrimSpace(slug))
	if isASCIISlug(s) {
		return s
	}
	t := strings.ToLower(strings.TrimSpace(title))
	if isASCIISlug(t) {
		return t
	}
	return "g-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
}

func isASCIISlug(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			return false
		}
	}
	return true
}

func (s *Service) CreateCatalogSkill(ctx context.Context, groupID uuid.UUID, title string) (*domain.Skill, error) {
	if s.skills == nil {
		return nil, domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	title = strings.TrimSpace(title)
	if title == "" || groupID == uuid.Nil {
		return nil, domain.Invalid("گروه و عنوان مهارت الزامی است")
	}
	if _, err := s.skills.GetGroup(ctx, groupID); err != nil {
		return nil, domain.Invalid("گروه مهارت نامعتبر است")
	}
	if existing, err := s.skills.GetSkillByTitle(ctx, groupID, title); err == nil {
		return existing, nil
	}
	sk := &domain.Skill{ID: uuid.New(), GroupID: groupID, Title: title, Status: "active"}
	if err := s.skills.CreateSkill(ctx, sk); err != nil {
		return nil, err
	}
	return sk, nil
}

func (s *Service) UpdateCatalogSkill(ctx context.Context, id uuid.UUID, title, status string, groupID uuid.UUID) (*domain.Skill, error) {
	if s.skills == nil {
		return nil, domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	sk, err := s.skills.GetSkill(ctx, id)
	if err != nil {
		return nil, err
	}
	if t := strings.TrimSpace(title); t != "" {
		sk.Title = t
	}
	if status != "" {
		sk.Status = status
	}
	if groupID != uuid.Nil {
		if _, err := s.skills.GetGroup(ctx, groupID); err != nil {
			return nil, domain.Invalid("گروه مهارت نامعتبر است")
		}
		sk.GroupID = groupID
	}
	if err := s.skills.UpdateSkill(ctx, sk); err != nil {
		return nil, err
	}
	return sk, nil
}

func (s *Service) UpdateGroup(ctx context.Context, id uuid.UUID, title string, sortOrder int) (*domain.SkillGroup, error) {
	if s.skills == nil {
		return nil, domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	g, err := s.skills.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	if t := strings.TrimSpace(title); t != "" {
		g.Title = t
	}
	if sortOrder != 0 {
		g.SortOrder = sortOrder
	}
	if err := s.skills.UpdateGroup(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	if s.skills == nil {
		return domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	if _, err := s.skills.GetGroup(ctx, id); err != nil {
		return err
	}
	return s.skills.DeleteGroup(ctx, id)
}

func (s *Service) DeleteCatalogSkill(ctx context.Context, id uuid.UUID) error {
	if s.skills == nil {
		return domain.Invalid("کاتالوگ مهارت در دسترس نیست")
	}
	if _, err := s.skills.GetSkill(ctx, id); err != nil {
		return err
	}
	return s.skills.DeleteSkill(ctx, id)
}
