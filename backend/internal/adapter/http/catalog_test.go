package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCatalogCoversEveryRegisteredRoute(t *testing.T) {
	h := NewRouter(Deps{})
	mux, ok := h.(chi.Routes)
	if !ok {
		t.Fatalf("router is %T, want chi.Routes", h)
	}

	type key struct{ method, path string }
	registered := map[key]struct{}{}
	if err := chi.Walk(mux, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		route = strings.ReplaceAll(route, "/*", "/")
		registered[key{method, strings.TrimSpace(route)}] = struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	catalog := map[key]CatalogRoute{}
	for _, r := range CatalogRoutes() {
		catalog[key{r.Method, r.Path}] = r
	}

	var missing []string
	for k := range registered {
		if _, ok := catalog[k]; ok {
			continue
		}
		// chi may emit "/api/v1/admin/skills" without the trailing slash used in Mount.
		alt := k
		if strings.HasSuffix(k.path, "/") {
			alt.path = strings.TrimSuffix(k.path, "/")
		} else {
			alt.path = k.path + "/"
		}
		if _, ok := catalog[alt]; ok {
			continue
		}
		missing = append(missing, k.method+" "+k.path)
	}
	var extra []string
	for k, r := range catalog {
		if _, ok := registered[k]; ok {
			continue
		}
		alt := k
		if strings.HasSuffix(k.path, "/") {
			alt.path = strings.TrimSuffix(k.path, "/")
		} else {
			alt.path = k.path + "/"
		}
		if _, ok := registered[alt]; ok {
			continue
		}
		extra = append(extra, r.Method+" "+r.Path)
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		t.Errorf("router routes missing from catalog:\n  %s", strings.Join(missing, "\n  "))
	}
	if len(extra) > 0 {
		t.Errorf("catalog routes not registered:\n  %s", strings.Join(extra, "\n  "))
	}
}

func TestAPICatalogJSONListsAllGroups(t *testing.T) {
	h := NewRouter(Deps{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, needle := range []string{
		`"service":"mahak-volunteer-api"`,
		`"id":"operator"`,
		"/api/v1/webhooks/events",
		"/api/v1/tickets",
		"/api/v1/admin/assignments/{id}/revision",
		"/api/v1/certificates/requests",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("catalog JSON missing %s", needle)
		}
	}
}

func TestOpenAPICoversCanonicalCatalog(t *testing.T) {
	root := findRepoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "docs", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]struct{}{}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "  /") || !strings.HasSuffix(line, ":") {
			continue
		}
		p := strings.TrimSuffix(strings.TrimSpace(line), ":")
		paths[p] = struct{}{}
		paths[strings.TrimSuffix(p, "/")] = struct{}{}
		if !strings.HasSuffix(p, "/") {
			paths[p+"/"] = struct{}{}
		}
	}
	var missing []string
	for _, r := range CatalogRoutes() {
		if r.Alias {
			continue
		}
		if _, ok := paths[r.Path]; !ok {
			missing = append(missing, r.Method+" "+r.Path)
		}
	}
	if len(missing) > 0 {
		t.Errorf("canonical catalog paths missing from docs/openapi.yaml:\n  %s", strings.Join(missing, "\n  "))
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "docs", "openapi.yaml")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("repo root not found")
	return ""
}

func TestMissionVerifyTokenIgnoresInternalBearer(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/events", nil)
	req.Header.Set("Authorization", "Bearer internal-secret")
	if got := missionVerifyToken(req, "internal-secret", ""); got != "" {
		t.Fatalf("internal bearer leaked as mission token: %q", got)
	}
	req.Header.Set("X-Mission-Token", "mission-secret")
	if got := missionVerifyToken(req, "internal-secret", ""); got != "mission-secret" {
		t.Fatalf("got %q", got)
	}
	if got := missionVerifyToken(req, "internal-secret", "body-secret"); got != "body-secret" {
		t.Fatalf("body should win, got %q", got)
	}
}
