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
	Auth          *authuc.Service
	Volunteers    *volunteeruc.Service
	Tasks         *taskuc.Service
	Missions      *missionuc.Service
	Certs         *certuc.Service
	Users         domain.UserRepository
	Stats         domain.StatsRepository
	Notify        domain.NotificationRepository
	Ready         func() map[string]string
	InternalKey   string
	WebhookSecret string
	CORSOrigins   []string
	Production    bool
}

func NewRouter(d Deps) http.Handler {
	origins := d.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	authLimit := newRateLimiter(20, time.Minute)
	apiLimit := newRateLimiter(300, time.Minute)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Compress(5))
	r.Use(securityHeaders)
	r.Use(maxBody(6 << 20))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Internal-Key", "X-Webhook-Secret"},
		ExposedHeaders:   []string{"Link", "Content-Disposition"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "mahak-volunteers"})
	})
	r.Get("/readyz", d.readyz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(apiLimit.middleware)
		r.With(authLimit.middleware).Post("/auth/register", d.register)
		r.With(authLimit.middleware).Post("/auth/login", d.login)
		if d.InternalKey != "" {
			r.With(requireSecret("X-Internal-Key", d.InternalKey)).Post("/auth/external", d.external)
		} else if !d.Production {
			r.Post("/auth/external", d.external)
		} else {
			r.Post("/auth/external", func(w http.ResponseWriter, _ *http.Request) {
				writeError(w, domain.ErrUnauthorized)
			})
		}
		r.Get("/certificates/verify/{code}", d.verifyCert)
		r.Get("/certificates/{code}/pdf", d.certPDF)
		if d.WebhookSecret != "" {
			r.With(requireSecret("X-Webhook-Secret", d.WebhookSecret)).Post("/webhooks/events", d.webhook)
		} else if !d.Production {
			r.Post("/webhooks/events", d.webhook)
		} else {
			r.Post("/webhooks/events", func(w http.ResponseWriter, _ *http.Request) {
				writeError(w, domain.ErrUnauthorized)
			})
		}

		r.Group(func(r chi.Router) {
			r.Use(d.authMiddleware)
			r.Get("/me", d.me)
			r.Get("/notifications", d.notifications)
			r.Post("/notifications/{id}/read", d.markRead)

			r.Get("/volunteers/me", d.myProfile)
			r.Put("/volunteers/me", d.updateProfile)
			r.Post("/volunteers/me/submit", d.submitProfile)
			r.Put("/volunteers/me/availability", d.setAvailability)
			r.Get("/volunteers/me/availability", d.myAvailability)
			r.Post("/volunteers/me/documents", d.uploadDoc)
			r.Get("/volunteers/me/documents", d.myDocs)

			r.Get("/tasks", d.listEligibleTasks)
			r.Get("/tasks/{id}", d.getTask)
			r.Post("/tasks/{id}/accept", d.acceptTask)
			r.Get("/assignments/me", d.myAssignments)
			r.Post("/assignments/{id}/rate", d.rateAssignment)
			r.Post("/assignments/{id}/cancel", d.cancelMine)

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

func (d Deps) readyz(w http.ResponseWriter, _ *http.Request) {
	checks := map[string]string{"status": "ready"}
	if d.Ready != nil {
		checks = d.Ready()
	}
	status := http.StatusOK
	if checks["status"] != "ready" {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, checks)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal"
	msg := "خطای داخلی سرور"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, msg = http.StatusNotFound, "not_found", "مورد یافت نشد"
	case errors.Is(err, domain.ErrUnauthorized):
		status, code, msg = http.StatusUnauthorized, "unauthorized", "ورود لازم است"
	case errors.Is(err, domain.ErrForbidden):
		status, code, msg = http.StatusForbidden, "forbidden", "دسترسی غیرمجاز"
	case errors.Is(err, domain.ErrInvalidInput):
		status, code, msg = http.StatusBadRequest, "invalid_input", "ورودی نامعتبر است"
	case errors.Is(err, domain.ErrInvalidTransition):
		status, code, msg = http.StatusBadRequest, "invalid_transition", "این تغییر وضعیت مجاز نیست"
	case errors.Is(err, domain.ErrDocumentRequired):
		status, code, msg = http.StatusBadRequest, "document_required", "مدارک الزامی ناقص است"
	case errors.Is(err, domain.ErrInvalidFileType):
		status, code, msg = http.StatusBadRequest, "invalid_file", "نوع فایل مجاز نیست"
	case errors.Is(err, domain.ErrFileTooLarge):
		status, code, msg = http.StatusBadRequest, "file_too_large", "حجم فایل بیش از حد مجاز است"
	case errors.Is(err, domain.ErrCertificateNotReady):
		status, code, msg = http.StatusBadRequest, "certificate_not_ready", "گواهی هنوز قابل صدور نیست"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyAssigned):
		status, code, msg = http.StatusConflict, "conflict", "این مورد قبلا ثبت شده است"
	case errors.Is(err, domain.ErrBusy):
		status, code, msg = http.StatusConflict, "busy", "سرور مشغول است، دوباره تلاش کنید"
	case errors.Is(err, domain.ErrTooManyRequests):
		status, code, msg = http.StatusTooManyRequests, "rate_limited", "تعداد درخواست‌ها بیش از حد است"
	case errors.Is(err, domain.ErrCapacityFull):
		status, code, msg = http.StatusUnprocessableEntity, "capacity_full", "ظرفیت تسک تکمیل است"
	case errors.Is(err, domain.ErrNotEligible):
		status, code, msg = http.StatusUnprocessableEntity, "not_eligible", "واجد شرایط این تسک نیستید"
	case errors.Is(err, domain.ErrNotApproved):
		status, code, msg = http.StatusUnprocessableEntity, "not_approved", "حساب داوطلبی شما تایید نشده است"
	case errors.Is(err, domain.ErrMissionExpired):
		status, code, msg = http.StatusUnprocessableEntity, "mission_expired", "مهلت ماموریت گذشته است"
	}
	writeJSON(w, status, map[string]string{"error": msg, "code": code})
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
