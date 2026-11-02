package terms

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

var (
	// testConfig is the configuration of the handlers, replaced by setup.
	testConfig *config.Configuration

	// testStorage is the storage of the handlers, reset at the end of each test.
	testStorage = &fakeStorage{}
)

func setup(t *testing.T) string {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"status.terms.accepted": "accepted", "placeholder": "placeholder"})

	dir := t.TempDir()
	testConfig = &config.Configuration{}
	testConfig.Terms.Version = "2"
	testConfig.Terms.Path = dir
	t.Cleanup(func() { *testStorage = fakeStorage{} })

	return dir
}

// fakeStorage is a Storage whose methods call the function set by the test, if any.
type fakeStorage struct {
	update func(ctx context.Context, userID, version string) error
	latest func(ctx context.Context, userID string) (string, error)
}

func (f *fakeStorage) UpdateVersionOfTermsAcceptedByUser(ctx context.Context, userID, version string) error {
	if f.update == nil {
		return nil
	}

	return f.update(ctx, userID, version)
}

func (f *fakeStorage) GetLatestVersionOfTermsAcceptedByUser(ctx context.Context, userID string) (string, error) {
	if f.latest == nil {
		return "", nil
	}

	return f.latest(ctx, userID)
}

func newRouter(handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(ctxkey.Config.Inject(testConfig), Inject(testStorage), ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
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
	ctxkey.Config.Set(c, testConfig)
	SetStorage(c, testStorage)
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
		update     func(context.Context, string, string) error
		wantStatus int
	}{
		{"missing version", `{}`, nil, http.StatusBadRequest},
		{"outdated version", `{"version":"1"}`, nil, http.StatusBadRequest},
		{"update fails", `{"version":"2"}`, func(context.Context, string, string) error { return errors.New("boom") }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStorage.update = tt.update

			req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			if code := apitest.Serve(newRouter(AcceptTerms), req).Code; code != tt.wantStatus {
				t.Errorf("status = %d, want %d", code, tt.wantStatus)
			}
		})
	}

	t.Run("accepted without storage", func(t *testing.T) {
		router := gin.New()
		router.Use(ctxkey.Config.Inject(testConfig), ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
		router.Any("/", AcceptTerms)

		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"version":"2"}`))
		if code := apitest.Serve(router, req).Code; code != http.StatusOK {
			t.Errorf("status = %d, want %d", code, http.StatusOK)
		}
	})

	t.Run("accepted and saved", func(t *testing.T) {
		var savedVersion string
		testStorage.update = func(_ context.Context, _, version string) error {
			savedVersion = version
			return nil
		}

		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"version":"2"}`))
		if code := apitest.Serve(newRouter(AcceptTerms), req).Code; code != http.StatusOK || savedVersion != "2" {
			t.Errorf("status = %d, saved version = %q, want 200 and 2", code, savedVersion)
		}
	})
}
