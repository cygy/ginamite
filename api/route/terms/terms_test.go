package terms

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/ctxlocalization"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"

	"github.com/gin-gonic/gin"
)

var (
	// testConfig is the configuration of the handlers, replaced by setup.
	testConfig *config.Configuration

	// testFunctions are the functions of the handlers, reset at the end of each test.
	testFunctions Functions
)

func setup(t *testing.T) string {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"status.terms.accepted": "accepted", "placeholder": "placeholder"})

	dir := t.TempDir()
	testConfig = &config.Configuration{}
	testConfig.Terms.Version = "2"
	testConfig.Terms.Path = dir
	t.Cleanup(func() { testFunctions = Functions{} })

	return dir
}

func newRouter(handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(ctxconfig.Inject(testConfig), Inject(&testFunctions), ctxlocalization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
	router.Any("/", handler)

	return router
}

func TestGetTerms(t *testing.T) {
	dir := setup(t)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	if code := apitest.Serve(newRouter(GetTerms()), req).Code; code != http.StatusInternalServerError {
		t.Errorf("missing terms file: status = %d, want %d", code, http.StatusInternalServerError)
	}

	if err := os.MkdirAll(filepath.Join(dir, "2"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2", "en.txt"), []byte("the terms"), 0o600); err != nil {
		t.Fatal(err)
	}

	// The file is read once, then served from the cache.
	router := newRouter(GetTerms())
	for range 2 {
		recorder := apitest.Serve(router, req)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "the terms") {
			t.Fatalf("status = %d, body = %q, want 200 with the terms", recorder.Code, recorder.Body.String())
		}
		if err := os.Remove(filepath.Join(dir, "2", "en.txt")); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
}

func TestGetTermsVersion(t *testing.T) {
	setup(t)

	c, recorder := apitest.NewContext("GET", "/", nil)
	ctxconfig.Set(c, testConfig)
	SetFunctions(c, &testFunctions)
	GetTermsVersion(c)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"version":"2"`) {
		t.Errorf("status = %d, body = %q, want 200 with version 2", recorder.Code, recorder.Body.String())
	}
}

func TestAcceptTerms(t *testing.T) {
	setup(t)

	tests := []struct {
		name       string
		body       string
		update     UpdateVersionOfTermsAcceptedByUserFunc
		wantStatus int
	}{
		{"missing version", `{}`, nil, http.StatusBadRequest},
		{"outdated version", `{"version":"1"}`, nil, http.StatusBadRequest},
		{"accepted without storage", `{"version":"2"}`, nil, http.StatusOK},
		{"update fails", `{"version":"2"}`, func(*gin.Context, string, string) error { return errors.New("boom") }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFunctions.UpdateVersionOfTermsAcceptedByUser = tt.update

			req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			if code := apitest.Serve(newRouter(AcceptTerms), req).Code; code != tt.wantStatus {
				t.Errorf("status = %d, want %d", code, tt.wantStatus)
			}
		})
	}

	t.Run("accepted and saved", func(t *testing.T) {
		var savedVersion string
		testFunctions.UpdateVersionOfTermsAcceptedByUser = func(_ *gin.Context, _, version string) error {
			savedVersion = version
			return nil
		}

		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"version":"2"}`))
		if code := apitest.Serve(newRouter(AcceptTerms), req).Code; code != http.StatusOK || savedVersion != "2" {
			t.Errorf("status = %d, saved version = %q, want 200 and 2", code, savedVersion)
		}
	})
}
