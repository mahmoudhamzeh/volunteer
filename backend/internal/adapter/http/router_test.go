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
			r.Put("/groups/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Delete("/groups/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Put("/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Delete("/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		})
		r.Route("/admin/skill-catalog", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Post("/groups", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
			r.Put("/groups/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Delete("/groups/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		})
	})

	gid := "11111111-1111-1111-1111-111111111111"
	for _, tc := range []struct {
		method, url string
		want        int
	}{
		{http.MethodPost, "/api/v1/admin/skills/groups", http.StatusCreated},
		{http.MethodPut, "/api/v1/admin/skills/groups/" + gid, http.StatusOK},
		{http.MethodDelete, "/api/v1/admin/skills/groups/" + gid, http.StatusOK},
		{http.MethodPost, "/api/v1/admin/skill-catalog/groups", http.StatusCreated},
		{http.MethodPut, "/api/v1/admin/skill-catalog/groups/" + gid, http.StatusOK},
		{http.MethodDelete, "/api/v1/admin/skill-catalog/groups/" + gid, http.StatusOK},
	} {
		req := httptest.NewRequest(tc.method, tc.url, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("%s %s status=%d body=%s", tc.method, tc.url, rec.Code, rec.Body.String())
		}
	}
}
