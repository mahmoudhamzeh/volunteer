package domain

import (
	"time"

	"github.com/google/uuid"
)

type SkillGroup struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	Skills    []Skill   `json:"skills"`
}

type Skill struct {
	ID         uuid.UUID `json:"id"`
	GroupID    uuid.UUID `json:"group_id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	GroupTitle string    `json:"group_title,omitempty"`
}

type VolunteerSkill struct {
	SkillID    uuid.UUID `json:"skill_id"`
	Title      string    `json:"title"`
	GroupID    uuid.UUID `json:"group_id"`
	GroupSlug  string    `json:"group_slug"`
	GroupTitle string    `json:"group_title"`
}

type SkillProposalStatus string

const (
	ProposalPending  SkillProposalStatus = "pending"
	ProposalApproved SkillProposalStatus = "approved"
	ProposalRejected SkillProposalStatus = "rejected"
)

type SkillProposal struct {
	ID             uuid.UUID           `json:"id"`
	VolunteerID    uuid.UUID           `json:"volunteer_id"`
	VolunteerName  string              `json:"volunteer_name,omitempty"`
	GroupID        uuid.UUID           `json:"group_id"`
	GroupTitle     string              `json:"group_title"`
	Title          string              `json:"title"`
	Status         SkillProposalStatus `json:"status"`
	AdminNote      string              `json:"admin_note"`
	CreatedSkillID uuid.UUID           `json:"created_skill_id,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
}
