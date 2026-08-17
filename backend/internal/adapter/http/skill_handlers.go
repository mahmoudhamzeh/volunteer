package httpserver

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/volunteeruc"
)

func (d Deps) skillCatalog(w http.ResponseWriter, r *http.Request) {
	items, err := d.Volunteers.Catalog(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) proposeSkill(w http.ResponseWriter, r *http.Request) {
	var in struct {
		GroupID string `json:"group_id"`
		Title   string `json:"title"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	gid, err := uuid.Parse(in.GroupID)
	if err != nil {
		writeError(w, domain.Invalid("گروه مهارت را انتخاب کنید"))
		return
	}
	p, err := d.Volunteers.ProposeSkill(r.Context(), mustPrincipal(r).ID, gid, in.Title)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (d Deps) mySkillProposals(w http.ResponseWriter, r *http.Request) {
	items, err := d.Volunteers.MyProposals(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) adminSkillProposals(w http.ResponseWriter, r *http.Request) {
	status := domain.SkillProposalStatus(r.URL.Query().Get("status"))
	items, err := d.Volunteers.ListProposals(r.Context(), status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) reviewSkillProposal(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in volunteeruc.ProposalReviewInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	p, err := d.Volunteers.ReviewProposal(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (d Deps) createSkillGroup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Slug      string `json:"slug"`
		Title     string `json:"title"`
		SortOrder int    `json:"sort_order"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	g, err := d.Volunteers.CreateGroup(r.Context(), in.Slug, in.Title, in.SortOrder)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (d Deps) createCatalogSkill(w http.ResponseWriter, r *http.Request) {
	var in struct {
		GroupID string `json:"group_id"`
		Title   string `json:"title"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	gid, err := uuid.Parse(in.GroupID)
	if err != nil {
		writeError(w, domain.Invalid("گروه مهارت نامعتبر است"))
		return
	}
	sk, err := d.Volunteers.CreateCatalogSkill(r.Context(), gid, in.Title)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sk)
}

func (d Deps) updateCatalogSkill(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Title   string `json:"title"`
		Status  string `json:"status"`
		GroupID string `json:"group_id"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var gid uuid.UUID
	if in.GroupID != "" {
		gid, err = uuid.Parse(in.GroupID)
		if err != nil {
			writeError(w, domain.Invalid("گروه مهارت نامعتبر است"))
			return
		}
	}
	sk, err := d.Volunteers.UpdateCatalogSkill(r.Context(), id, in.Title, in.Status, gid)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sk)
}
