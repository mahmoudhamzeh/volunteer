package httpserver

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/volunteeruc"
)

func contextWithPrincipal(ctx context.Context, p principal) context.Context {
	return context.WithValue(ctx, ctxUser, p)
}

func mustPrincipal(r *http.Request) principal {
	p, _ := r.Context().Value(ctxUser).(principal)
	return p
}

type creds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

func (d Deps) register(w http.ResponseWriter, r *http.Request) {
	var in creds
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	u, token, err := d.Auth.Register(r.Context(), in.Email, in.Password, in.FullName, domain.Role(in.Role))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": userDTO(u)})
}

func (d Deps) login(w http.ResponseWriter, r *http.Request) {
	var in creds
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	u, token, err := d.Auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": userDTO(u)})
}

func (d Deps) external(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ExternalUserID string `json:"external_user_id"`
		Email          string `json:"email"`
		FullName       string `json:"full_name"`
		Role           string `json:"role"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	u, token, err := d.Auth.UpsertFromExternal(r.Context(), in.ExternalUserID, in.Email, in.FullName, domain.Role(in.Role))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": userDTO(u)})
}

func (d Deps) me(w http.ResponseWriter, r *http.Request) {
	p := mustPrincipal(r)
	u, err := d.Users.GetByID(r.Context(), p.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := map[string]any{"user": userDTO(u)}
	if u.Role == domain.RoleVolunteer {
		if v, err := d.Volunteers.GetMine(r.Context(), u.ID); err == nil {
			resp["volunteer"] = volunteerDTO(v)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (d Deps) myProfile(w http.ResponseWriter, r *http.Request) {
	v, err := d.Volunteers.GetMine(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, volunteerDTO(v))
}

func (d Deps) updateProfile(w http.ResponseWriter, r *http.Request) {
	var in volunteeruc.ProfileInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	v, err := d.Volunteers.UpsertProfile(r.Context(), mustPrincipal(r).ID, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, volunteerDTO(v))
}

func (d Deps) submitProfile(w http.ResponseWriter, r *http.Request) {
	v, err := d.Volunteers.SubmitForReview(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, volunteerDTO(v))
}

func (d Deps) setAvailability(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Slots []domain.AvailabilitySlot `json:"slots"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if err := d.Volunteers.SetAvailability(r.Context(), mustPrincipal(r).ID, in.Slots); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (d Deps) myAvailability(w http.ResponseWriter, r *http.Request) {
	v, err := d.Volunteers.GetMine(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	slots, err := d.Volunteers.ListAvailability(r.Context(), v.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(slots))
}

func (d Deps) uploadDoc(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	defer file.Close()
	kind := domain.DocumentKind(r.FormValue("kind"))
	if kind == "" {
		kind = domain.DocOther
	}
	mime := hdr.Header.Get("Content-Type")
	doc, err := d.Volunteers.UploadDocument(r.Context(), mustPrincipal(r).ID, kind, hdr.Filename, mime, hdr.Size, file)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (d Deps) myDocs(w http.ResponseWriter, r *http.Request) {
	v, err := d.Volunteers.GetMine(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	docs, err := d.Volunteers.ListDocuments(r.Context(), v.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(docs))
}

func (d Deps) listEligibleTasks(w http.ResponseWriter, r *http.Request) {
	f := domain.TaskFilter{Query: r.URL.Query().Get("q"), Skill: domain.SkillCategory(r.URL.Query().Get("skill")), Limit: queryInt(r, "limit", 50), Offset: queryInt(r, "offset", 0)}
	items, total, err := d.Tasks.ListEligible(r.Context(), mustPrincipal(r).ID, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonempty(items), "total": total})
}

func (d Deps) getTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	p := mustPrincipal(r)
	if p.Role == domain.RoleVolunteer {
		v, err := d.Volunteers.GetMine(r.Context(), p.ID)
		if err != nil {
			writeError(w, err)
			return
		}
		if !v.Status.CanViewTasks() {
			writeError(w, domain.ErrNotApproved)
			return
		}
	}
	t, err := d.Tasks.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (d Deps) acceptTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	a, err := d.Tasks.Accept(r.Context(), mustPrincipal(r).ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (d Deps) myAssignments(w http.ResponseWriter, r *http.Request) {
	items, err := d.Tasks.MyAssignments(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) rateAssignment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	a, err := d.Tasks.RateByVolunteer(r.Context(), mustPrincipal(r).ID, id, in.Rating, in.Comment)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (d Deps) volunteerCancel(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	a, err := d.Tasks.CancelByOwner(r.Context(), mustPrincipal(r).ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (d Deps) listMissions(w http.ResponseWriter, r *http.Request) {
	items, err := d.Missions.List(r.Context(), true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) startMission(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	p, err := d.Missions.Start(r.Context(), mustPrincipal(r).ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (d Deps) missionProgress(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Increment int `json:"increment"`
	}
	_ = decodeJSON(r, &in)
	if in.Increment <= 0 {
		in.Increment = 1
	}
	p, err := d.Missions.ReportProgress(r.Context(), mustPrincipal(r).ID, id, in.Increment)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (d Deps) myMissions(w http.ResponseWriter, r *http.Request) {
	items, err := d.Missions.MyProgress(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) myCerts(w http.ResponseWriter, r *http.Request) {
	items, err := d.Certs.ListMine(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) verifyCert(w http.ResponseWriter, r *http.Request) {
	code, err := parseID(r, "code")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	c, err := d.Certs.Verify(r.Context(), code)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, certDTO(c))
}

func (d Deps) certPDF(w http.ResponseWriter, r *http.Request) {
	code, err := parseID(r, "code")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	pdf, c, err := d.Certs.PDF(r.Context(), code)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=mahak-certificate-"+c.VerificationCode.String()+".pdf")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

func (d Deps) webhook(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Event       string `json:"event"`
		VolunteerID string `json:"volunteer_id"`
		Increment   int    `json:"increment"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	vid, err := uuid.Parse(in.VolunteerID)
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if err := d.Missions.AwardWebhook(r.Context(), vid, in.Event, in.Increment); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "ok"})
}

func (d Deps) notifications(w http.ResponseWriter, r *http.Request) {
	items, err := d.Notify.ListByUser(r.Context(), mustPrincipal(r).ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) markRead(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if err := d.Notify.MarkRead(r.Context(), id, mustPrincipal(r).ID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (d Deps) dashboard(w http.ResponseWriter, r *http.Request) {
	s, err := d.Stats.Dashboard(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (d Deps) adminVolunteers(w http.ResponseWriter, r *http.Request) {
	f := domain.VolunteerFilter{
		Status: domain.VolunteerStatus(r.URL.Query().Get("status")),
		Skill:  domain.SkillCategory(r.URL.Query().Get("skill")),
		Query:  r.URL.Query().Get("q"),
		Limit:  queryInt(r, "limit", 20),
		Offset: queryInt(r, "offset", 0),
	}
	items, total, err := d.Volunteers.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]any, 0, len(items))
	for i := range items {
		out = append(out, volunteerDTO(&items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "total": total})
}

func (d Deps) adminVolunteer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	v, err := d.Volunteers.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	docs, _ := d.Volunteers.ListDocuments(r.Context(), v.ID)
	slots, _ := d.Volunteers.ListAvailability(r.Context(), v.ID)
	writeJSON(w, http.StatusOK, map[string]any{"volunteer": volunteerDTO(v), "documents": nonempty(docs), "availability": nonempty(slots)})
}

func (d Deps) reviewVolunteer(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	v, err := d.Volunteers.Review(r.Context(), mustPrincipal(r).ID, id, in.Action, in.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, volunteerDTO(v))
}

func (d Deps) adminDocs(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	docs, err := d.Volunteers.ListDocuments(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

func (d Deps) streamDoc(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	rc, doc, err := d.Volunteers.OpenDocument(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rc.Close()
	safe := filepath.Base(strings.ReplaceAll(doc.FileName, `"`, ""))
	if safe == "" || safe == "." || safe == string(filepath.Separator) {
		safe = "document"
	}
	w.Header().Set("Content-Type", doc.MimeType)
	w.Header().Set("Content-Disposition", `inline; filename="`+safe+`"`)
	_, _ = io.Copy(w, rc)
}

func (d Deps) adminAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	slots, err := d.Volunteers.ListAvailability(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, slots)
}

type taskBody struct {
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Location          string    `json:"location"`
	StartsAt          time.Time `json:"starts_at"`
	EndsAt            time.Time `json:"ends_at"`
	Capacity          int       `json:"capacity"`
	HourWeight        float64   `json:"hour_weight"`
	RequiredSkills    []string  `json:"required_skills"`
	MinScore          float64   `json:"min_score"`
	RequiredEducation string    `json:"required_education"`
	Status            string    `json:"status"`
}

func (d Deps) adminTasks(w http.ResponseWriter, r *http.Request) {
	f := domain.TaskFilter{Query: r.URL.Query().Get("q"), Status: domain.TaskStatus(r.URL.Query().Get("status")), Limit: queryInt(r, "limit", 50), Offset: queryInt(r, "offset", 0)}
	items, total, err := d.Tasks.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonempty(items), "total": total})
}

func (d Deps) createTask(w http.ResponseWriter, r *http.Request) {
	var in taskBody
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	t, err := d.Tasks.Create(r.Context(), mustPrincipal(r).ID, taskInput(in))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (d Deps) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in taskBody
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	t, err := d.Tasks.Update(r.Context(), id, taskInput(in))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (d Deps) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	if err := d.Tasks.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (d Deps) adminAssignments(w http.ResponseWriter, r *http.Request) {
	f := domain.AssignmentFilter{Status: domain.AssignmentStatus(r.URL.Query().Get("status")), Limit: queryInt(r, "limit", 50), Offset: queryInt(r, "offset", 0)}
	if vid := r.URL.Query().Get("volunteer_id"); vid != "" {
		if id, err := uuid.Parse(vid); err == nil {
			f.VolunteerID = id
		}
	}
	if tid := r.URL.Query().Get("task_id"); tid != "" {
		if id, err := uuid.Parse(tid); err == nil {
			f.TaskID = id
		}
	}
	items, total, err := d.Tasks.ListAssignments(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonempty(items), "total": total})
}

func (d Deps) attendance(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	a, err := d.Tasks.ConfirmAttendance(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (d Deps) complete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Discipline int    `json:"discipline"`
		Expertise  int    `json:"expertise"`
		Ethics     int    `json:"ethics"`
		Comment    string `json:"comment"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	a, err := d.Tasks.Complete(r.Context(), id, in.Discipline, in.Expertise, in.Ethics, in.Comment)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (d Deps) cancelAssignment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	a, err := d.Tasks.Cancel(r.Context(), id, true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (d Deps) issueCert(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	c, err := d.Certs.IssueForAssignment(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (d Deps) issueAggregated(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	from := time.Now().AddDate(-1, 0, 0)
	to := time.Now()
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t
		}
	}
	c, err := d.Certs.IssueAggregated(r.Context(), id, from, to)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (d Deps) adminMissions(w http.ResponseWriter, r *http.Request) {
	items, err := d.Missions.List(r.Context(), false)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) createMission(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		Kind          string  `json:"kind"`
		HourWeight    float64 `json:"hour_weight"`
		DeadlineHours *int    `json:"deadline_hours"`
		WebhookEvent  string  `json:"webhook_event"`
		TargetCount   int     `json:"target_count"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	m, err := d.Missions.Create(r.Context(), missionIn(in.Title, in.Description, in.Kind, in.HourWeight, in.DeadlineHours, in.WebhookEvent, in.TargetCount))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (d Deps) updateMission(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	var in struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		Kind          string  `json:"kind"`
		HourWeight    float64 `json:"hour_weight"`
		DeadlineHours *int    `json:"deadline_hours"`
		WebhookEvent  string  `json:"webhook_event"`
		TargetCount   int     `json:"target_count"`
		Status        string  `json:"status"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}
	mi := missionIn(in.Title, in.Description, in.Kind, in.HourWeight, in.DeadlineHours, in.WebhookEvent, in.TargetCount)
	mi.Status = domain.MissionStatus(in.Status)
	m, err := d.Missions.Update(r.Context(), id, mi)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (d Deps) ranking(w http.ResponseWriter, r *http.Request) {
	items, err := d.Stats.Ranking(r.Context(), queryInt(r, "limit", 20))
	if err != nil {
		writeError(w, err)
		return
	}
	if r.URL.Query().Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=mahak-ranking.csv")
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"rank", "name", "city", "hours", "score", "tasks"})
		for i, row := range items {
			_ = cw.Write([]string{
				strconv.Itoa(i + 1), row.FullName, row.City,
				strconv.FormatFloat(row.TotalHours, 'f', 1, 64),
				strconv.FormatFloat(row.AverageScore, 'f', 2, 64),
				strconv.Itoa(row.CompletedTasks),
			})
		}
		cw.Flush()
		return
	}
	writeJSON(w, http.StatusOK, nonempty(items))
}

func (d Deps) skills(w http.ResponseWriter, r *http.Request) {
	m, err := d.Stats.SkillDistribution(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (d Deps) apiCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"product":    "Mahak Volunteer Management Platform",
		"product_fa": "سامانه مدیریت داوطلبان محک",
		"service":    "mahak-volunteer-api",
		"version":    "v1",
		"auth":       "Authorization: Bearer <jwt>  |  X-Internal-Token for integrations",
		"groups": []map[string]any{
			{"name": "health", "items": []string{"GET /healthz", "GET /readyz", "GET /api/v1"}},
			{"name": "auth", "items": []string{"POST /api/v1/auth/register", "POST /api/v1/auth/login", "POST /api/v1/auth/external"}},
			{"name": "session", "items": []string{"GET /api/v1/me", "GET /api/v1/notifications", "POST /api/v1/notifications/{id}/read"}},
			{"name": "volunteer_profile", "items": []string{
				"GET /api/v1/volunteers/me", "PUT /api/v1/volunteers/me", "POST /api/v1/volunteers/me/submit",
				"GET /api/v1/volunteers/me/availability", "PUT /api/v1/volunteers/me/availability",
				"GET /api/v1/volunteers/me/documents", "POST /api/v1/volunteers/me/documents",
			}},
			{"name": "volunteer_work", "items": []string{
				"GET /api/v1/tasks", "GET /api/v1/tasks/{id}", "POST /api/v1/tasks/{id}/accept",
				"GET /api/v1/assignments/me", "POST /api/v1/assignments/{id}/rate", "POST /api/v1/assignments/{id}/cancel",
				"GET /api/v1/missions", "POST /api/v1/missions/{id}/start", "POST /api/v1/missions/{id}/progress", "GET /api/v1/missions/me",
				"GET /api/v1/certificates/me",
			}},
			{"name": "public_certificates", "items": []string{"GET /api/v1/certificates/verify/{code}", "GET /api/v1/certificates/{code}/pdf"}},
			{"name": "admin_volunteers", "items": []string{
				"GET /api/v1/admin/dashboard", "GET /api/v1/admin/volunteers", "GET /api/v1/admin/volunteers/{id}",
				"POST /api/v1/admin/volunteers/{id}/review", "GET /api/v1/admin/volunteers/{id}/documents",
				"GET /api/v1/admin/volunteers/{id}/availability", "GET /api/v1/admin/documents/{id}",
				"POST /api/v1/admin/volunteers/{id}/certificates/aggregated",
			}},
			{"name": "admin_operations", "items": []string{
				"GET /api/v1/admin/tasks", "POST /api/v1/admin/tasks", "PUT /api/v1/admin/tasks/{id}", "DELETE /api/v1/admin/tasks/{id}",
				"GET /api/v1/admin/assignments", "POST /api/v1/admin/assignments/{id}/attendance",
				"POST /api/v1/admin/assignments/{id}/complete", "POST /api/v1/admin/assignments/{id}/cancel",
				"POST /api/v1/admin/assignments/{id}/certificate",
				"GET /api/v1/admin/missions", "POST /api/v1/admin/missions", "PUT /api/v1/admin/missions/{id}",
				"GET /api/v1/admin/reports/ranking", "GET /api/v1/admin/reports/skills",
			}},
			{"name": "integrations", "items": []string{"POST /api/v1/webhooks/events"}},
		},
	})
}
