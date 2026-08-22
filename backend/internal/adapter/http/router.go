package httpserver

import (
	"context"
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
	Auth          *authuc.Service
	Volunteers    *volunteeruc.Service
	Tasks         *taskuc.Service
	Missions      *missionuc.Service
	Certs         *certuc.Service
	Users         domain.UserRepository
	Stats         domain.StatsRepository
	Notify        domain.NotificationRepository
	Storage       domain.ObjectStorage
	InternalToken string
	CORSOrigins   []string
	Ready         func(context.Context) error
}

func NewRouter(d Deps) http.Handler {
	origins := d.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestSize(8 << 20))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Internal-Token", "X-Request-Id"},
		ExposedHeaders:   []string{"Link", "Content-Disposition", "X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "mahak-volunteer-api"})
	})
	r.Get("/readyz", d.readyz)
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "این مسیر در API وجود ندارد"})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "این روش برای این مسیر مجاز نیست"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", d.apiCatalog)
		r.Post("/auth/register", d.register)
		r.Post("/auth/login", d.login)
		r.Post("/auth/otp/send", d.sendOTP)
		r.Post("/auth/otp/verify", d.verifyOTP)
		r.With(d.requireInternalToken).Post("/auth/external", d.external)
		r.Get("/certificates/verify/{code}", d.verifyCert)
		r.Get("/certificates/{code}/pdf", d.certPDF)
		r.With(d.requireInternalToken).Post("/webhooks/events", d.webhook)

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
			r.Delete("/volunteers/me/documents/{id}", d.deleteMyDoc)
			r.Post("/volunteers/me/skill-proposals", d.proposeSkill)
			r.Get("/volunteers/me/skill-proposals", d.mySkillProposals)

			r.Get("/tasks", d.listEligibleTasks)
			r.Get("/tasks/{id}", d.getTask)
			r.Post("/tasks/{id}/accept", d.acceptTask)
			r.Get("/assignments/me", d.myAssignments)
			r.Post("/assignments/{id}/rate", d.rateAssignment)
			r.Post("/assignments/{id}/cancel", d.cancelMyAssignment)
			r.Post("/assignments/{id}/start", d.startAssignment)
			r.Post("/assignments/{id}/deliver", d.deliverAssignment)

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
				r.Post("/admin/volunteers/{id}/status", d.setVolunteerStatus)
				r.Post("/admin/volunteers/{id}/comments", d.commentVolunteer)
				r.Put("/admin/volunteers/{id}", d.adminUpdateVolunteer)
				r.Get("/admin/volunteers/{id}/documents", d.adminDocs)
				r.Get("/admin/documents/{id}", d.streamDoc)
				r.Get("/admin/volunteers/{id}/availability", d.adminAvailability)

				r.Get("/admin/tasks", d.adminTasks)
				r.Post("/admin/tasks", d.createTask)
				r.Put("/admin/tasks/{id}", d.updateTask)
				r.Post("/admin/tasks/{id}/status", d.setTaskStatus)
				r.Post("/admin/tasks/{id}/assign", d.assignVolunteer)
				r.Delete("/admin/tasks/{id}", d.deleteTask)
				r.Get("/admin/tasks/{id}/assignments", d.adminTaskAssignments)

				r.Get("/admin/assignments", d.adminAssignments)
				r.Post("/admin/assignments/{id}/approve", d.approveAssignment)
				r.Post("/admin/assignments/{id}/reject", d.rejectAssignment)
				r.Post("/admin/assignments/{id}/message", d.messageAssignment)
				r.Post("/admin/assignments/{id}/attendance", d.attendance)
				r.Post("/admin/assignments/{id}/complete", d.complete)
				r.Post("/admin/assignments/{id}/cancel", d.cancelAssignment)
				r.Post("/admin/assignments/{id}/certificate", d.issueCert)
				r.Get("/admin/assignments/{id}/delivery", d.streamDelivery)

				r.Get("/admin/missions", d.adminMissions)
				r.Post("/admin/missions", d.createMission)
				r.Put("/admin/missions/{id}", d.updateMission)

				r.Post("/admin/volunteers/{id}/certificates/aggregated", d.issueAggregated)
				r.Get("/admin/reports/ranking", d.ranking)
				r.Get("/admin/reports/skills", d.skills)

				r.Route("/admin/skills", func(r chi.Router) {
					r.Get("/", d.skillCatalog)
					r.Get("/catalog", d.skillCatalog)
					r.Post("/groups", d.createSkillGroup)
					r.Put("/groups/{id}", d.updateSkillGroup)
					r.Delete("/groups/{id}", d.deleteSkillGroup)
					r.Post("/", d.createCatalogSkill)
					r.Put("/{id}", d.updateCatalogSkill)
					r.Delete("/{id}", d.deleteCatalogSkill)
					r.Get("/proposals", d.adminSkillProposals)
					r.Post("/proposals/{id}/review", d.reviewSkillProposal)
				})
				r.Route("/admin/skill-catalog", func(r chi.Router) {
					r.Get("/", d.skillCatalog)
					r.Post("/groups", d.createSkillGroup)
					r.Put("/groups/{id}", d.updateSkillGroup)
					r.Delete("/groups/{id}", d.deleteSkillGroup)
					r.Post("/skills", d.createCatalogSkill)
					r.Put("/skills/{id}", d.updateCatalogSkill)
					r.Delete("/skills/{id}", d.deleteCatalogSkill)
				})
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

func (d Deps) requireInternalToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(d.InternalToken) == "" {
			writeError(w, domain.ErrUnauthorized)
			return
		}
		got := strings.TrimSpace(r.Header.Get("X-Internal-Token"))
		if got == "" {
			h := r.Header.Get("Authorization")
			if strings.HasPrefix(h, "Bearer ") {
				got = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			}
		}
		if got != d.InternalToken {
			writeError(w, domain.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (d Deps) readyz(w http.ResponseWriter, r *http.Request) {
	if d.Ready != nil {
		if err := d.Ready(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "service": "mahak-volunteer-api"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "service": "mahak-volunteer-api"})
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
		errors.Is(err, domain.ErrNotApproved), errors.Is(err, domain.ErrMissionExpired),
		errors.Is(err, domain.ErrMissionNotVerified):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrBusy):
		status = http.StatusServiceUnavailable
	}
	low := strings.ToLower(msg)
	if strings.Contains(low, "readonly") || strings.Contains(low, "read only") || strings.Contains(low, "read-only") {
		msg = "سامانه موقتاً در حال به‌روزرسانی است؛ چند لحظه بعد دوباره تلاش کنید"
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
