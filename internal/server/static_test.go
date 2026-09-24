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

	router := gin.New()
	registerStaticFiles(router, dir)

	page := httptest.NewRecorder()
	router.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	if page.Code != http.StatusOK || page.Body.String() != "<h1>My app</h1>" {
		t.Fatalf("static page = (%d, %q), want (200, page)", page.Code, page.Body.String())
	}

	api := httptest.NewRecorder()
	router.ServeHTTP(api, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	if api.Code != http.StatusNotFound {
		t.Fatalf("unknown API status = %d, want %d", api.Code, http.StatusNotFound)
	}
}
