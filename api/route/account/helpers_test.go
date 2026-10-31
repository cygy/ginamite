package account

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

var errBoom = errors.New("boom")

const (
	testSecret  = "test-secret"
	testTokenID = "token-1"
	testUserID  = "42"
)

var testUser = authentication.User{
	ID:           testUserID,
	Username:     "john",
	EmailAddress: "john@example.com",
	Locale:       "en",
	PrivateKey:   "private-key",
}

// testCase is a scenario of a handler: the request, the injected functions and the expected HTTP status.
type testCase struct {
	name   string
	id     string // value of the "id" path parameter.
	key    string // value of the "key" path parameter.
	body   string // JSON body.
	userID string // if not empty, the request is authenticated as this user.
	setup  func()
	want   int
	check  func(t *testing.T, body string)
}

// testConfig is the configuration of the handlers, replaced by setup.
var testConfig *config.Configuration

// setup initializes the global dependencies of the handlers and resets every injected function at the end of the test.
func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	testConfig = config.NewConfiguration()
	testConfig.SupportedLocales = []string{"en"}
	testConfig.JWT.Secret = testSecret
	testConfig.Terms.Version = "v1"
	testConfig.Environment = environment.Development
	timezone.Main = timezone.NewConfiguration()
	timezone.Main.Timezones = []string{"Europe/Paris"}
	kafka.Main = kafka.NewClient()

	authentication.BuildTokenFromID = func(context.Context, string) (*authentication.Token, error) {
		return &authentication.Token{UserID: testUserID, ExpirationDate: time.Now().Add(time.Hour), CreationDate: time.Now()}, nil
	}
	authentication.ExtraPropertiesForTokenWithID = func(string) (map[string]string, error) { return nil, nil }
	authentication.ExtendTokenExpirationDateFromID = func(context.Context, string, time.Duration) error { return nil }

	originalIDValidator := request.IsValidID
	request.IsValidID = func(id string) bool { return len(id) == len(validNotificationID) }

	t.Cleanup(func() {
		request.IsValidID = originalIDValidator
		resetFunctions()
		authentication.BuildTokenFromID = nil
		authentication.ExtraPropertiesForTokenWithID = nil
		authentication.ExtendTokenExpirationDateFromID = nil
	})
}

// testFunctions are the functions of the handlers, set by the tests and reset at the end of each one.
var testFunctions Functions

// resetFunctions removes every function of the handlers.
func resetFunctions() { testFunctions = Functions{} }

// newRouter returns a router with the locale and the auth token middlewares, serving the handler on every path shape.
func newRouter(handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(ctxconfig.Inject(testConfig), Inject(&testFunctions), ctxLocale.Inject("en", []string{"en"}))
	router.Use(ctxAuthToken.Parse(testSecret, nil, false))
	router.Any("/", handler)
	router.Any("/:id", handler)
	router.Any("/:id/:key", handler)

	return router
}

// do sends the request of the test case to the handler and returns the response.
func do(t *testing.T, handler gin.HandlerFunc, tc testCase) *httptest.ResponseRecorder {
	t.Helper()

	target := "/"
	if tc.id != "" {
		target += tc.id
		if tc.key != "" {
			target += "/" + tc.key
		}
	}

	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(tc.body))
	if tc.userID != "" {
		req.Header.Set("Authorization", "Bearer "+apitest.SignToken(t, testSecret, tc.userID, testTokenID, false))
	}

	return apitest.Serve(newRouter(handler), req)
}

// runCases runs the scenarios against the handler.
func runCases(t *testing.T, handler gin.HandlerFunc, tests []testCase) {
	t.Helper()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)
			if tc.setup != nil {
				tc.setup()
			}

			recorder := do(t, handler, tc)

			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, tc.want, recorder.Body.String())
			}
			if tc.check != nil {
				tc.check(t, recorder.Body.String())
			}
		})
	}
}

// newContext returns a context carrying the "en" locale and the path parameters.
func newContext(params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	c, recorder := apitest.NewContext(http.MethodPost, "/", nil)
	c.Params = params
	ctxconfig.Set(c, testConfig)
	SetFunctions(c, &testFunctions)
	ctxLocale.Set(c, "en", "en")

	return c, recorder
}

// all combines several setup functions.
func all(fns ...func()) func() {
	return func() {
		for _, fn := range fns {
			fn()
		}
	}
}

// inTestEnvironment switches the application to the test environment.
func inTestEnvironment() { testConfig.Environment = environment.Test }

// withUser defines GetUserDetailsByID.
func withUser(user authentication.User, err error) func() {
	return func() {
		testFunctions.GetUserDetailsByID = func(*gin.Context, string) (authentication.User, error) { return user, err }
	}
}

// contains returns a check verifying the body contains every string.
func contains(values ...string) func(*testing.T, string) {
	return func(t *testing.T, body string) {
		t.Helper()

		for _, v := range values {
			if !strings.Contains(body, v) {
				t.Errorf("body %q must contain %q", body, v)
			}
		}
	}
}

// fakeValidator is a TokenValidator returning the configured token informations.
type fakeValidator struct {
	source string
	infos  validator.TokenInfos
	ok     bool
}

func (v fakeValidator) SourceName() string { return v.source }

func (v fakeValidator) VerifyTokenAndExtractInfos(string) (validator.TokenInfos, bool) {
	return v.infos, v.ok
}

// withValidator makes the handlers use the fake validator for every third-party method.
func withValidator(v fakeValidator) func() {
	return func() {
		testFunctions.NewTokenValidator = func(*gin.Context, string) validator.TokenValidator { return v }
	}
}
