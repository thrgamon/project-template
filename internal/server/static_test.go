package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStaticFilesServeFrontendAndNotUnknownAPIPaths(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>My app</h1>"), 0o600); err != nil {
		t.Fatalf("write static index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dashboard.html"), []byte("<h1>Dashboard</h1>"), 0o600); err != nil {
		t.Fatalf("write static page: %v", err)
	}

	router := gin.New()
	registerStaticFiles(router, dir)

	page := httptest.NewRecorder()
	router.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	if page.Code != http.StatusOK || page.Body.String() != "<h1>My app</h1>" {
		t.Fatalf("static page = (%d, %q), want (200, page)", page.Code, page.Body.String())
	}

	// adapter-static writes /dashboard as dashboard.html; a direct link must load it.
	dashboard := httptest.NewRecorder()
	router.ServeHTTP(dashboard, httptest.NewRequest(http.MethodGet, "/dashboard", nil))
	if dashboard.Code != http.StatusOK || dashboard.Body.String() != "<h1>Dashboard</h1>" {
		t.Fatalf("prerendered page = (%d, %q), want (200, page)", dashboard.Code, dashboard.Body.String())
	}

	missing := httptest.NewRecorder()
	router.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing page status = %d, want %d", missing.Code, http.StatusNotFound)
	}

	api := httptest.NewRecorder()
	router.ServeHTTP(api, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	if api.Code != http.StatusNotFound {
		t.Fatalf("unknown API status = %d, want %d", api.Code, http.StatusNotFound)
	}
}
