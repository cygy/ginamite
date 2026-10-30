package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

// setup initializes the global dependencies of the handlers and resets the injected functions at the end of the test.
func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	config.Main = config.NewConfiguration()
	config.Main.SupportedLocales = []string{"en"}
	timezone.Main = timezone.NewConfiguration()
	timezone.Main.Timezones = []string{"Europe/Paris"}
	kafka.Main = kafka.NewClient()

	t.Cleanup(func() {
		GetEnabledUsersForAdmin, GetDisabledUsersForAdmin = nil, nil
		UpdatePasswordByAdmin, UpdateInformationsByAdmin, UpdateDescriptionByAdmin = nil, nil, nil
		UpdateSocialByAdmin, UpdateSettingsByAdmin, UpdateNotificationsByAdmin = nil, nil, nil
		GetTokensForAdmin, DeleteTokenByIDByAdmin = nil, nil
		SetAbilitiesToUser, AddAbilitiesToUser, RemoveAbilitiesFromUser = nil, nil, nil
		account.GetTokenDetailsByID, account.GetUserDetailsByID = nil, nil
		account.DeleteUserByID, account.DisableUserByID, account.EnableUserByID = nil, nil, nil
	})
}

// serve sends a request with the given JSON body to the handler, mounted on "/:id", and returns the HTTP status.
func serve(handler gin.HandlerFunc, body string) int {
	router := gin.New()
	router.Use(ctxLocale.Inject("en", []string{"en"}))
	router.Any("/:id", handler)

	req, _ := http.NewRequest(http.MethodPost, "/42", strings.NewReader(body))

	return apitest.Serve(router, req).Code
}

// newTestContext returns a context carrying the JSON body and the "en" locale.
func newTestContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	c, recorder := apitest.NewContext(http.MethodPost, "/42", strings.NewReader(body))
	ctxLocale.Set(c, "en", "en")

	return c, recorder
}
