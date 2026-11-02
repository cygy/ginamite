package account

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/cygy/ginamite/common/notifications"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/queue"
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

// testTimezones are the timezones of the handlers.
var testTimezones = &timezone.Configuration{Timezones: []string{"Europe/Paris"}}

// testIDValidator accepts the identifiers of the length of an ObjectID.
func testIDValidator(id string) bool { return len(id) == len(validNotificationID) }

// testQueue is the queue of the handlers: it has no producer, the messages are only logged.
var testQueue = queue.New(broker.NewMemory())

// setup initializes the global dependencies of the handlers and resets every injected function at the end of the test.
func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	testConfig = config.NewConfiguration()
	testConfig.SupportedLocales = []string{"en"}
	testConfig.JWT.Secret = testSecret
	testConfig.Terms.Version = "v1"
	testConfig.Environment = environment.Development

	resetFunctions()

	testHooks.BuildTokenFromID = func(context.Context, string) (*authentication.Token, error) {
		return &authentication.Token{UserID: testUserID, ExpirationDate: time.Now().Add(time.Hour), CreationDate: time.Now()}, nil
	}
	testHooks.ExtraPropertiesForTokenWithID = func(string) (map[string]string, error) { return nil, nil }
	testHooks.ExtendTokenExpirationDateFromID = func(context.Context, string, time.Duration) error { return nil }

	t.Cleanup(func() {
		resetFunctions()
		testHooks = authentication.Hooks{}
	})
}

var (
	// testFunctions are the functions of the handlers, set by the tests and reset at the end of each one.
	testFunctions Functions

	// testHooks are the functions building and extending the auth tokens, set by setup.
	testHooks authentication.Hooks
)

// testNotifications is the storage of the notifications of the handlers, reset at the end of each test.
var testNotifications = &fakeNotifications{}

// resetFunctions removes every function of the handlers and resets the storages.
func resetFunctions() {
	*testNotifications = fakeNotifications{}
	testFunctions = Functions{Notifications: testNotifications}
}

// withoutNotifications removes the storage of the notifications.
func withoutNotifications() { testFunctions.Notifications = nil }

// fakeNotifications is a NotificationsStorage whose methods call the function set by the test, if any.
type fakeNotifications struct {
	receivedFn    func(ctx context.Context, userID, locale string) (any, error)
	receivedOneFn func(ctx context.Context, userID, notificationID, locale string) (any, error)
	readFn        func(ctx context.Context, userID, notificationID string) error
	notifiedFn    func(ctx context.Context, userID, notificationID string) error
	deleteFn      func(ctx context.Context, userID, notificationID string) error
	subscribedFn  func(ctx context.Context, userID string) ([]string, error)
	subscribeFn   func(ctx context.Context, userID string, notifications []string) error
	unsubscribeFn func(ctx context.Context, userID, notification, unsubscribeKey string) error
}

func (f *fakeNotifications) NotificationsReceivedByUser(ctx context.Context, userID, locale string) (any, error) {
	if f.receivedFn == nil {
		return nil, nil
	}

	return f.receivedFn(ctx, userID, locale)
}

func (f *fakeNotifications) NotificationReceivedByUser(ctx context.Context, userID, notificationID, locale string) (any, error) {
	if f.receivedOneFn == nil {
		return nil, nil
	}

	return f.receivedOneFn(ctx, userID, notificationID, locale)
}

func (f *fakeNotifications) UpdateUserNotificationRead(ctx context.Context, userID, notificationID string) error {
	if f.readFn == nil {
		return nil
	}

	return f.readFn(ctx, userID, notificationID)
}

func (f *fakeNotifications) UpdateUserNotificationNotified(ctx context.Context, userID, notificationID string) error {
	if f.notifiedFn == nil {
		return nil
	}

	return f.notifiedFn(ctx, userID, notificationID)
}

func (f *fakeNotifications) DeleteUserNotification(ctx context.Context, userID, notificationID string) error {
	if f.deleteFn == nil {
		return nil
	}

	return f.deleteFn(ctx, userID, notificationID)
}

func (f *fakeNotifications) NotificationsSubscribedByUser(ctx context.Context, userID string) ([]string, error) {
	if f.subscribedFn == nil {
		return nil, nil
	}

	return f.subscribedFn(ctx, userID)
}

func (f *fakeNotifications) SubscribeNotificationsToUser(ctx context.Context, userID string, notifications []string) error {
	if f.subscribeFn == nil {
		return nil
	}

	return f.subscribeFn(ctx, userID, notifications)
}

func (f *fakeNotifications) UnsubscribeNotificationFromUser(ctx context.Context, userID, notification, unsubscribeKey string) error {
	if f.unsubscribeFn == nil {
		return nil
	}

	return f.unsubscribeFn(ctx, userID, notification, unsubscribeKey)
}

// newRouter returns a router with the locale and the auth token middlewares, serving the handler on every path shape.
func newRouter(handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(ctxkey.Config.Inject(testConfig), ctxkey.Queue.Inject(testQueue), request.InjectIDValidator(testIDValidator), ctxkey.Timezones.Inject(testTimezones), ctxkey.Notifications.Inject(notifications.NewRegistry()), Inject(&testFunctions), ctxAuthToken.InjectHooks(testHooks), ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
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
	ctxkey.Config.Set(c, testConfig)
	ctxkey.Queue.Set(c, testQueue)
	request.SetIDValidator(c, testIDValidator)
	ctxkey.Timezones.Set(c, testTimezones)
	ctxkey.Notifications.Set(c, notifications.NewRegistry())
	SetFunctions(c, &testFunctions)
	ctxAuthToken.SetHooks(c, testHooks)
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
