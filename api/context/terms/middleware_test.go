package terms

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/route/terms"

	"github.com/gin-gonic/gin"
)

// fakeStorage is a terms.Storage returning the latest version of the terms accepted by the users from a function.
type fakeStorage struct {
	latest func(ctx context.Context, userID string) (string, error)
}

func (f *fakeStorage) UpdateVersionOfTermsAcceptedByUser(context.Context, string, string) error {
	return nil
}

func (f *fakeStorage) GetLatestVersionOfTermsAcceptedByUser(ctx context.Context, userID string) (string, error) {
	return f.latest(ctx, userID)
}

func TestVerifyLatestVersionOfTermsAcceptedByUser(t *testing.T) {
	const secret = "secret"

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	latest := func(version string) *fakeStorage {
		return &fakeStorage{func(context.Context, string) (string, error) { return version, nil }}
	}

	tests := []struct {
		name       string
		authed     bool
		storage    *fakeStorage
		wantStatus int
	}{
		{"no storage is only reported", true, nil, http.StatusOK},
		{"storage fails", true, &fakeStorage{func(context.Context, string) (string, error) { return "", errors.New("boom") }}, http.StatusInternalServerError},
		{"anonymous user is ignored", false, latest("1"), http.StatusOK},
		{"latest version", true, latest("2"), http.StatusOK},
		{"old version is refused", true, latest("1"), http.StatusPreconditionFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
			router.Use(ctxAuthToken.Parse(secret, nil, false))
			if tt.storage != nil {
				router.Use(terms.Inject(tt.storage))
			}
			router.Use(VerifyLatestVersionOfTermsAcceptedByUser("2"))
			router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			if tt.authed {
				req.Header.Set("Authorization", "Bearer "+apitest.SignToken(t, secret, "user-1", "token-1", false))
			}

			if got := apitest.Serve(router, req).Code; got != tt.wantStatus {
				t.Errorf("status = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}
