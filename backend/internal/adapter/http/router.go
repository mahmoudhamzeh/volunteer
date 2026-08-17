package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/authuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/certuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/missionuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/taskuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/volunteeruc"
)

type Deps struct {
	Auth       *authuc.Service
	Volunteers *volunteeruc.Service
	Tasks      *taskuc.Service
	Missions   *missionuc.Service
	Certs      *certuc.Service
	Users      domain.UserRepository
	Stats      domain.StatsRepository
	Notify     domain.NotificationRepository
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link", "Content-Disposition"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "mahak-volunteers"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", d.register)
		r.Post("/auth/login", d.login)
		r.Post("/auth/external", d.external)
		r.Get("/certificates/verify/{code}", d.verifyCert)
		r.Get("/certificates/{code}/pdf", d.certPDF)
		r.Post("/webhooks/events", d.webhook)

		r.Group(func(r chi.Router) {
			r.Use(d.authMiddleware)
			r.Get("/me", d.me)
			r.Get("/notifications", d.notifications)
			r.Post("/notifications/{id}/read", d.markRead)

			r.Get("/skills", d.skillCatalog)
			r.Get("/volunteers/me", d.myProfile)
			r.Put("/volunteers/me", d.updateProfile)
			r.Post("/volunteers/me/submit", d.submitProfile)
			r.Put("/volunteers/me/availability", d.setAvailability)
			r.Get("/volunteers/me/availability", d.myAvailability)
			r.Post("/volunteers/me/documents", d.uploadDoc)
			r.Get("/volunteers/me/documents", d.myDocs)
			r.Post("/volunteers/me/skill-proposals", d.proposeSkill)
			r.Get("/volunteers/me/skill-proposals", d.mySkillProposals)

			r.Get("/tasks", d.listEligibleTasks)
			r.Get("/tasks/{id}", d.getTask)
			r.Post("/tasks/{id}/accept", d.acceptTask)
			r.Get("/assignments/me", d.myAssignments)
			r.Post("/assignments/{id}/rate", d.rateAssignment)

			r.Get("/missions", d.listMissions)
			r.Post("/missions/{id}/start", d.startMission)
			r.Post("/missions/{id}/progress", d.missionProgress)
			r.Get("/missions/me", d.myMissions)

			r.Get("/certificates/me", d.myCerts)

			r.Group(func(r chi.Router) {
				r.Use(d.staffOnly)
				r.Get("/admin/dashboard", d.dashboard)
				r.Get("/admin/volunteers", d.adminVolunteers)
				r.Get("/admin/volunteers/{id}", d.adminVolunteer)
				r.Post("/admin/volunteers/{id}/review", d.reviewVolunteer)
				r.Get("/admin/volunteers/{id}/documents", d.adminDocs)
				r.Get("/admin/documents/{id}", d.streamDoc)
				r.Get("/admin/volunteers/{id}/availability", d.adminAvailability)

				r.Get("/admin/tasks", d.adminTasks)
				r.Post("/admin/tasks", d.createTask)
				r.Put("/admin/tasks/{id}", d.updateTask)
				r.Delete("/admin/tasks/{id}", d.deleteTask)

				r.Get("/admin/assignments", d.adminAssignments)
				r.Post("/admin/assignments/{id}/attendance", d.attendance)
				r.Post("/admin/assignments/{id}/complete", d.complete)
				r.Post("/admin/assignments/{id}/cancel", d.cancelAssignment)
				r.Post("/admin/assignments/{id}/certificate", d.issueCert)

				r.Get("/admin/missions", d.adminMissions)
				r.Post("/admin/missions", d.createMission)
				r.Put("/admin/missions/{id}", d.updateMission)

				r.Post("/admin/volunteers/{id}/certificates/aggregated", d.issueAggregated)
				r.Get("/admin/reports/ranking", d.ranking)
				r.Get("/admin/reports/skills", d.skills)
				r.Get("/admin/skill-catalog", d.skillCatalog)
				r.Post("/admin/skill-catalog/groups", d.createSkillGroup)
				r.Post("/admin/skill-catalog/skills", d.createCatalogSkill)
				r.Put("/admin/skill-catalog/skills/{id}", d.updateCatalogSkill)
				r.Get("/admin/skill-proposals", d.adminSkillProposals)
				r.Post("/admin/skill-proposals/{id}/review", d.reviewSkillProposal)
			})
		})
	})
	return r
}

type ctxKey string

const ctxUser ctxKey = "user"

type principal struct {
	ID   uuid.UUID
	Role domain.Role
}

func (d Deps) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeError(w, domain.ErrUnauthorized)
			return
		}
		claims, err := d.Auth.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			writeError(w, domain.ErrUnauthorized)
			return
		}
		ctx := contextWithPrincipal(r.Context(), principal{ID: claims.UserID, Role: claims.Role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (d Deps) staffOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := mustPrincipal(r)
		if p.Role != domain.RoleAdmin && p.Role != domain.RoleOperator {
			writeError(w, domain.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrDocumentRequired), errors.Is(err, domain.ErrInvalidFileType),
		errors.Is(err, domain.ErrFileTooLarge), errors.Is(err, domain.ErrCertificateNotReady):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyAssigned):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrCapacityFull), errors.Is(err, domain.ErrNotEligible),
		errors.Is(err, domain.ErrNotApproved), errors.Is(err, domain.ErrMissionExpired):
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func queryInt(r *http.Request, key string, def int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return def
	}
	return n
}

func parseID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, name))
}
