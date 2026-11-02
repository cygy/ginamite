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

	resetStorage()

	testHooks.BuildTokenFromID = func(context.Context, string) (*authentication.Token, error) {
		return &authentication.Token{UserID: testUserID, ExpirationDate: time.Now().Add(time.Hour), CreationDate: time.Now()}, nil
	}
	testHooks.ExtraPropertiesForTokenWithID = func(string) (map[string]string, error) { return nil, nil }
	testHooks.ExtendTokenExpirationDateFromID = func(context.Context, string, time.Duration) error { return nil }

	t.Cleanup(func() {
		resetStorage()
		testHooks = authentication.Hooks{}
	})
}

var (
	// testStorage are the functions of the handlers, set by the tests and reset at the end of each one.
	testStorage Storage

	// testHooks are the functions building and extending the auth tokens, set by setup.
	testHooks authentication.Hooks
)

// testNotifications, testTokens, testProcesses and testAccounts are the storages of the handlers, reset at the end of each test.
var (
	testNotifications = &fakeNotifications{}
	testTokens        = &fakeTokens{}
	testProcesses     = &fakeProcesses{}
	testAccounts      = &fakeAccounts{}
)

// resetStorage removes every function of the handlers and resets the storages.
func resetStorage() {
	*testNotifications = fakeNotifications{}
	*testTokens = fakeTokens{}
	*testProcesses = fakeProcesses{}
	*testAccounts = fakeAccounts{}
	testStorage = Storage{Notifications: testNotifications, Tokens: testTokens, Processes: testProcesses, Accounts: testAccounts}
}

// withoutNotifications, withoutTokens, withoutProcesses and withoutAccounts remove the storages.
func withoutNotifications() { testStorage.Notifications = nil }
func withoutTokens()        { testStorage.Tokens = nil }
func withoutProcesses()     { testStorage.Processes = nil }
func withoutAccounts()      { testStorage.Accounts = nil }

// fakeTokens is a TokensStorage whose methods call the function set by the test, if any.
type fakeTokens struct {
	saveFn       func(ctx context.Context, details authentication.Details) (string, bool, error)
	expirationFn func(ctx context.Context, tokenID string) (time.Time, error)
	deleteFn     func(ctx context.Context, tokenID string) error
	detailsFn    func(ctx context.Context, tokenID string) (authentication.Token, error)
	ownedFn      func(ctx context.Context, userID string) (any, error)
	ownedOneFn   func(ctx context.Context, tokenID string) (any, string, error)
	updateFn     func(ctx context.Context, tokenID, name string, enableNotifications bool, notifications []string) error
}

func (f *fakeTokens) SaveAuthenticationToken(ctx context.Context, details authentication.Details) (string, bool, error) {
	if f.saveFn == nil {
		return "", false, nil
	}

	return f.saveFn(ctx, details)
}

func (f *fakeTokens) GetTokenExpirationDate(ctx context.Context, tokenID string) (time.Time, error) {
	if f.expirationFn == nil {
		return time.Time{}, errors.New("no expiration date")
	}

	return f.expirationFn(ctx, tokenID)
}

func (f *fakeTokens) DeleteTokenByID(ctx context.Context, tokenID string) error {
	if f.deleteFn == nil {
		return nil
	}

	return f.deleteFn(ctx, tokenID)
}

func (f *fakeTokens) GetTokenDetailsByID(ctx context.Context, tokenID string) (authentication.Token, error) {
	if f.detailsFn == nil {
		return authentication.Token{}, nil
	}

	return f.detailsFn(ctx, tokenID)
}

func (f *fakeTokens) GetOwnedTokens(ctx context.Context, userID string) (any, error) {
	if f.ownedFn == nil {
		return nil, nil
	}

	return f.ownedFn(ctx, userID)
}

func (f *fakeTokens) GetOwnedTokenByID(ctx context.Context, tokenID string) (any, string, error) {
	if f.ownedOneFn == nil {
		return nil, "", nil
	}

	return f.ownedOneFn(ctx, tokenID)
}

func (f *fakeTokens) UpdateTokenByID(ctx context.Context, tokenID, name string, enableNotifications bool, notifications []string) error {
	if f.updateFn == nil {
		return nil
	}

	return f.updateFn(ctx, tokenID, name, enableNotifications, notifications)
}

// fakeProcesses is a ProcessesStorage whose methods call the function set by the test, if any.
type fakeProcesses struct {
	NewForgotPasswordProcessForUserFn     func(ctx context.Context, userID, processPrivateKey string) (string, error)
	GetForgotPasswordProcessDetailsByIDFn func(ctx context.Context, processID string) (string, string, error)
	WrongAccessToForgotPasswordProcessFn  func(ctx context.Context, processID string)
	DeleteForgotPasswordProcessFn         func(ctx context.Context, processID string) error
	CreateNewDeleteAccountProcessFn       func(ctx context.Context, userID string) (string, string, error)
	CreateNewDisableAccountProcessFn      func(ctx context.Context, userID string) (string, string, error)
	CreateNewUpdatePasswordProcessFn      func(ctx context.Context, userID, encryptedPassword string) (string, string, error)
	CreateNewUpdateEmailAddressProcessFn  func(ctx context.Context, userID, emailAddress string) (string, string, error)
	CreateNewVerifyEmailAddressProcessFn  func(ctx context.Context, userID string) (string, string, error)
	VerifyProcessFn                       func(ctx context.Context, processID, key string) (string, string, error)
	DeleteProcessByIDFn                   func(ctx context.Context, processID string) error
}

func (f *fakeProcesses) NewForgotPasswordProcessForUser(ctx context.Context, userID, processPrivateKey string) (string, error) {
	if f.NewForgotPasswordProcessForUserFn == nil {
		return "", nil
	}

	return f.NewForgotPasswordProcessForUserFn(ctx, userID, processPrivateKey)
}

func (f *fakeProcesses) GetForgotPasswordProcessDetailsByID(ctx context.Context, processID string) (string, string, error) {
	if f.GetForgotPasswordProcessDetailsByIDFn == nil {
		return "", "", nil
	}

	return f.GetForgotPasswordProcessDetailsByIDFn(ctx, processID)
}

func (f *fakeProcesses) WrongAccessToForgotPasswordProcess(ctx context.Context, processID string) {
	if f.WrongAccessToForgotPasswordProcessFn != nil {
		f.WrongAccessToForgotPasswordProcessFn(ctx, processID)
	}
}

func (f *fakeProcesses) DeleteForgotPasswordProcess(ctx context.Context, processID string) error {
	if f.DeleteForgotPasswordProcessFn == nil {
		return nil
	}

	return f.DeleteForgotPasswordProcessFn(ctx, processID)
}

func (f *fakeProcesses) CreateNewDeleteAccountProcess(ctx context.Context, userID string) (string, string, error) {
	if f.CreateNewDeleteAccountProcessFn == nil {
		return "", "", nil
	}

	return f.CreateNewDeleteAccountProcessFn(ctx, userID)
}

func (f *fakeProcesses) CreateNewDisableAccountProcess(ctx context.Context, userID string) (string, string, error) {
	if f.CreateNewDisableAccountProcessFn == nil {
		return "", "", nil
	}

	return f.CreateNewDisableAccountProcessFn(ctx, userID)
}

func (f *fakeProcesses) CreateNewUpdatePasswordProcess(ctx context.Context, userID, encryptedPassword string) (string, string, error) {
	if f.CreateNewUpdatePasswordProcessFn == nil {
		return "", "", nil
	}

	return f.CreateNewUpdatePasswordProcessFn(ctx, userID, encryptedPassword)
}

func (f *fakeProcesses) CreateNewUpdateEmailAddressProcess(ctx context.Context, userID, emailAddress string) (string, string, error) {
	if f.CreateNewUpdateEmailAddressProcessFn == nil {
		return "", "", nil
	}

	return f.CreateNewUpdateEmailAddressProcessFn(ctx, userID, emailAddress)
}

func (f *fakeProcesses) CreateNewVerifyEmailAddressProcess(ctx context.Context, userID string) (string, string, error) {
	if f.CreateNewVerifyEmailAddressProcessFn == nil {
		return "", "", nil
	}

	return f.CreateNewVerifyEmailAddressProcessFn(ctx, userID)
}

func (f *fakeProcesses) VerifyProcess(ctx context.Context, processID, key string) (string, string, error) {
	if f.VerifyProcessFn == nil {
		return "", "", nil
	}

	return f.VerifyProcessFn(ctx, processID, key)
}

func (f *fakeProcesses) DeleteProcessByID(ctx context.Context, processID string) error {
	if f.DeleteProcessByIDFn == nil {
		return nil
	}

	return f.DeleteProcessByIDFn(ctx, processID)
}

// fakeAccounts is an AccountsStorage whose methods call the function set by the test, if any.
type fakeAccounts struct {
	GetUserDetailsByIdentifierFn                                    func(ctx context.Context, identifier string) (authentication.User, error)
	GetUserDetailsByIDFn                                            func(ctx context.Context, userID string) (authentication.User, error)
	GetUserDetailsByFacebookUserIDFn                                func(ctx context.Context, facebookUserID string) (authentication.User, error)
	GetUserDetailsByGoogleUserIDFn                                  func(ctx context.Context, googleUserID string) (authentication.User, error)
	GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFn func(ctx context.Context, identifier string) (authentication.User, string, error)
	UpdateEncryptedPasswordByUserIDFn                               func(ctx context.Context, userID, encryptedPassword string) error
	UpdateUserPasswordFn                                            func(ctx context.Context, userID, password string, isPasswordEncrypted bool) error
	UpdateUserEmailAddressFn                                        func(ctx context.Context, userID, emailAddress string) error
	UserEmailAddressIsVerifiedFn                                    func(ctx context.Context, userID string) error
	UpdateUserSocialNetworksFn                                      func(ctx context.Context, userID, facebook, twitter, instagram string) error
	DoesAccountWithUsernameExistFn                                  func(ctx context.Context, username string) (bool, error)
	DoesAccountWithEmailAddressExistFn                              func(ctx context.Context, emailAddress string) (bool, error)
	DoesUserWithFacebookUserIDExistFn                               func(ctx context.Context, facebookUserID string) (bool, error)
	DoesUserWithGoogleUserIDExistFn                                 func(ctx context.Context, googleUserID string) (bool, error)
	RegisterByEmailAddressFn                                        func(ctx context.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error
	RegisterByThirdPartyTokenFn                                     func(ctx context.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error
	ValidateUserRegistrationFn                                      func(ctx context.Context, userID, registrationKey string) error
	CancelUserRegistrationFn                                        func(ctx context.Context, userID, registrationKey string) error
	GetRegistrationDetailsFn                                        func(ctx context.Context, userID string) (*authentication.Details, error)
	SaveFacebookUserTokenDetailsToUserFn                            func(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error
	SaveGoogleUserTokenDetailsToUserFn                              func(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error
	GetOwnedAccountDetailsFn                                        func(ctx context.Context, userID string) (any, error)
	UpdatedAccountInfosFn                                           func(ctx context.Context, userID, firstName, lastName string) error
	UpdatedAccountSettingsFn                                        func(ctx context.Context, userID, locale, timezone string) error
	DeleteUserByIDFn                                                func(ctx context.Context, userID string) error
	DisableUserByIDFn                                               func(ctx context.Context, userID string) (authentication.User, error)
	EnableUserByIDFn                                                func(ctx context.Context, userID string) (authentication.User, error)
}

func (f *fakeAccounts) GetUserDetailsByIdentifier(ctx context.Context, identifier string) (authentication.User, error) {
	if f.GetUserDetailsByIdentifierFn == nil {
		return authentication.User{}, nil
	}

	return f.GetUserDetailsByIdentifierFn(ctx, identifier)
}

func (f *fakeAccounts) GetUserDetailsByID(ctx context.Context, userID string) (authentication.User, error) {
	if f.GetUserDetailsByIDFn == nil {
		return authentication.User{}, nil
	}

	return f.GetUserDetailsByIDFn(ctx, userID)
}

func (f *fakeAccounts) GetUserDetailsByFacebookUserID(ctx context.Context, facebookUserID string) (authentication.User, error) {
	if f.GetUserDetailsByFacebookUserIDFn == nil {
		return authentication.User{}, nil
	}

	return f.GetUserDetailsByFacebookUserIDFn(ctx, facebookUserID)
}

func (f *fakeAccounts) GetUserDetailsByGoogleUserID(ctx context.Context, googleUserID string) (authentication.User, error) {
	if f.GetUserDetailsByGoogleUserIDFn == nil {
		return authentication.User{}, nil
	}

	return f.GetUserDetailsByGoogleUserIDFn(ctx, googleUserID)
}

func (f *fakeAccounts) GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(ctx context.Context, identifier string) (authentication.User, string, error) {
	if f.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFn == nil {
		return authentication.User{}, "", nil
	}

	return f.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFn(ctx, identifier)
}

func (f *fakeAccounts) UpdateEncryptedPasswordByUserID(ctx context.Context, userID, encryptedPassword string) error {
	if f.UpdateEncryptedPasswordByUserIDFn == nil {
		return nil
	}

	return f.UpdateEncryptedPasswordByUserIDFn(ctx, userID, encryptedPassword)
}

func (f *fakeAccounts) UpdateUserPassword(ctx context.Context, userID, password string, isPasswordEncrypted bool) error {
	if f.UpdateUserPasswordFn == nil {
		return nil
	}

	return f.UpdateUserPasswordFn(ctx, userID, password, isPasswordEncrypted)
}

func (f *fakeAccounts) UpdateUserEmailAddress(ctx context.Context, userID, emailAddress string) error {
	if f.UpdateUserEmailAddressFn == nil {
		return nil
	}

	return f.UpdateUserEmailAddressFn(ctx, userID, emailAddress)
}

func (f *fakeAccounts) UserEmailAddressIsVerified(ctx context.Context, userID string) error {
	if f.UserEmailAddressIsVerifiedFn == nil {
		return nil
	}

	return f.UserEmailAddressIsVerifiedFn(ctx, userID)
}

func (f *fakeAccounts) UpdateUserSocialNetworks(ctx context.Context, userID, facebook, twitter, instagram string) error {
	if f.UpdateUserSocialNetworksFn == nil {
		return nil
	}

	return f.UpdateUserSocialNetworksFn(ctx, userID, facebook, twitter, instagram)
}

func (f *fakeAccounts) DoesAccountWithUsernameExist(ctx context.Context, username string) (bool, error) {
	if f.DoesAccountWithUsernameExistFn == nil {
		return false, nil
	}

	return f.DoesAccountWithUsernameExistFn(ctx, username)
}

func (f *fakeAccounts) DoesAccountWithEmailAddressExist(ctx context.Context, emailAddress string) (bool, error) {
	if f.DoesAccountWithEmailAddressExistFn == nil {
		return false, nil
	}

	return f.DoesAccountWithEmailAddressExistFn(ctx, emailAddress)
}

func (f *fakeAccounts) DoesUserWithFacebookUserIDExist(ctx context.Context, facebookUserID string) (bool, error) {
	if f.DoesUserWithFacebookUserIDExistFn == nil {
		return false, nil
	}

	return f.DoesUserWithFacebookUserIDExistFn(ctx, facebookUserID)
}

func (f *fakeAccounts) DoesUserWithGoogleUserIDExist(ctx context.Context, googleUserID string) (bool, error) {
	if f.DoesUserWithGoogleUserIDExistFn == nil {
		return false, nil
	}

	return f.DoesUserWithGoogleUserIDExistFn(ctx, googleUserID)
}

func (f *fakeAccounts) RegisterByEmailAddress(ctx context.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	if f.RegisterByEmailAddressFn == nil {
		return nil
	}

	return f.RegisterByEmailAddressFn(ctx, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip, device)
}

func (f *fakeAccounts) RegisterByThirdPartyToken(ctx context.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	if f.RegisterByThirdPartyTokenFn == nil {
		return nil
	}

	return f.RegisterByThirdPartyTokenFn(ctx, username, tokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip, device)
}

func (f *fakeAccounts) ValidateUserRegistration(ctx context.Context, userID, registrationKey string) error {
	if f.ValidateUserRegistrationFn == nil {
		return nil
	}

	return f.ValidateUserRegistrationFn(ctx, userID, registrationKey)
}

func (f *fakeAccounts) CancelUserRegistration(ctx context.Context, userID, registrationKey string) error {
	if f.CancelUserRegistrationFn == nil {
		return nil
	}

	return f.CancelUserRegistrationFn(ctx, userID, registrationKey)
}

func (f *fakeAccounts) GetRegistrationDetails(ctx context.Context, userID string) (*authentication.Details, error) {
	if f.GetRegistrationDetailsFn == nil {
		return nil, nil
	}

	return f.GetRegistrationDetailsFn(ctx, userID)
}

func (f *fakeAccounts) SaveFacebookUserTokenDetailsToUser(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error {
	if f.SaveFacebookUserTokenDetailsToUserFn == nil {
		return nil
	}

	return f.SaveFacebookUserTokenDetailsToUserFn(ctx, userID, tokenInfos)
}

func (f *fakeAccounts) SaveGoogleUserTokenDetailsToUser(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error {
	if f.SaveGoogleUserTokenDetailsToUserFn == nil {
		return nil
	}

	return f.SaveGoogleUserTokenDetailsToUserFn(ctx, userID, tokenInfos)
}

func (f *fakeAccounts) GetOwnedAccountDetails(ctx context.Context, userID string) (any, error) {
	if f.GetOwnedAccountDetailsFn == nil {
		return nil, nil
	}

	return f.GetOwnedAccountDetailsFn(ctx, userID)
}

func (f *fakeAccounts) UpdatedAccountInfos(ctx context.Context, userID, firstName, lastName string) error {
	if f.UpdatedAccountInfosFn == nil {
		return nil
	}

	return f.UpdatedAccountInfosFn(ctx, userID, firstName, lastName)
}

func (f *fakeAccounts) UpdatedAccountSettings(ctx context.Context, userID, locale, timezone string) error {
	if f.UpdatedAccountSettingsFn == nil {
		return nil
	}

	return f.UpdatedAccountSettingsFn(ctx, userID, locale, timezone)
}

func (f *fakeAccounts) DeleteUserByID(ctx context.Context, userID string) error {
	if f.DeleteUserByIDFn == nil {
		return nil
	}

	return f.DeleteUserByIDFn(ctx, userID)
}

func (f *fakeAccounts) DisableUserByID(ctx context.Context, userID string) (authentication.User, error) {
	if f.DisableUserByIDFn == nil {
		return authentication.User{}, nil
	}

	return f.DisableUserByIDFn(ctx, userID)
}

func (f *fakeAccounts) EnableUserByID(ctx context.Context, userID string) (authentication.User, error) {
	if f.EnableUserByIDFn == nil {
		return authentication.User{}, nil
	}

	return f.EnableUserByIDFn(ctx, userID)
}

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
	router.Use(ctxkey.Config.Inject(testConfig), ctxkey.Queue.Inject(testQueue), request.InjectIDValidator(testIDValidator), ctxkey.Timezones.Inject(testTimezones), ctxkey.Notifications.Inject(notifications.NewRegistry()), Inject(&testStorage), ctxAuthToken.InjectHooks(testHooks), ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
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
	SetStorage(c, &testStorage)
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
		testAccounts.GetUserDetailsByIDFn = func(context.Context, string) (authentication.User, error) { return user, err }
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
		testStorage.NewTokenValidator = func(*gin.Context, string) validator.TokenValidator { return v }
	}
}
