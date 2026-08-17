package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSkillCatalogSubpathIsReachable(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/admin/reports/skills", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		r.Route("/admin/skills", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Post("/groups", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
		})
		r.Route("/admin/skill-catalog", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Post("/groups", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
		})
	})

	for _, u := range []string{"/api/v1/admin/skills/groups", "/api/v1/admin/skill-catalog/groups"} {
		req := httptest.NewRequest(http.MethodPost, u, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("POST %s status=%d body=%s", u, rec.Code, rec.Body.String())
		}
	}
}
