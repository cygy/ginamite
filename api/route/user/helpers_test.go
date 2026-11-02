package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/cygy/ginamite/common/notifications"

	"github.com/cygy/ginamite/common/queue"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

var (
	// testConfig is the configuration of the handlers, replaced by setup.
	testConfig *config.Configuration

	// testTimezones are the timezones of the handlers.
	testTimezones = &timezone.Configuration{Timezones: []string{"Europe/Paris"}}

	// testQueue is the queue of the handlers: it has no producer, the messages are only logged.
	testQueue = queue.New(broker.NewMemory())

	// testStorage is the storage of the handlers and testAccountFunctions their account functions, reset at the end of
	// each test. The storage is removed from the context by withoutStorage.
	testStorage          = &fakeStorage{}
	testAccountFunctions account.Storage
	storageRemoved       bool
	testTokens           = &fakeAccountTokens{}
	testAccounts         = &fakeAccountAccounts{}
)

// withoutStorage makes the handlers run as an application which provides no storage of the users.
func withoutStorage() { storageRemoved = true }

// fakeStorage is a Storage whose methods call the function set by the test, if any.
type fakeStorage struct {
	GetEnabledUsersForAdminFn    func(context.Context) (any, error)
	GetDisabledUsersForAdminFn   func(context.Context) (any, error)
	UpdatePasswordByAdminFn      func(context.Context, string, string) error
	UpdateInformationsByAdminFn  func(context.Context, string, string, string, string, bool) error
	UpdateDescriptionByAdminFn   func(context.Context, string, string, string) error
	UpdateSocialByAdminFn        func(context.Context, string, string, string, string) error
	UpdateSettingsByAdminFn      func(context.Context, string, string, string) error
	UpdateNotificationsByAdminFn func(context.Context, string, []string) error
	GetTokensForAdminFn          func(context.Context, string) (any, error)
	DeleteTokenByIDByAdminFn     func(context.Context, string) error
	SetAbilitiesToUserFn         func(context.Context, string, []string) error
	AddAbilitiesToUserFn         func(context.Context, string, []string) error
	RemoveAbilitiesFromUserFn    func(context.Context, string, []string) error
}

func (f *fakeStorage) GetEnabledUsersForAdmin(ctx context.Context) (any, error) {
	if f.GetEnabledUsersForAdminFn == nil {
		return nil, nil
	}

	return f.GetEnabledUsersForAdminFn(ctx)
}

func (f *fakeStorage) GetDisabledUsersForAdmin(ctx context.Context) (any, error) {
	if f.GetDisabledUsersForAdminFn == nil {
		return nil, nil
	}

	return f.GetDisabledUsersForAdminFn(ctx)
}

func (f *fakeStorage) UpdatePasswordByAdmin(ctx context.Context, userID, password string) error {
	if f.UpdatePasswordByAdminFn == nil {
		return nil
	}

	return f.UpdatePasswordByAdminFn(ctx, userID, password)
}

func (f *fakeStorage) UpdateInformationsByAdmin(ctx context.Context, userID, firstName, lastName, emailAddress string, validEmailAddress bool) error {
	if f.UpdateInformationsByAdminFn == nil {
		return nil
	}

	return f.UpdateInformationsByAdminFn(ctx, userID, firstName, lastName, emailAddress, validEmailAddress)
}

func (f *fakeStorage) UpdateDescriptionByAdmin(ctx context.Context, userID, locale, description string) error {
	if f.UpdateDescriptionByAdminFn == nil {
		return nil
	}

	return f.UpdateDescriptionByAdminFn(ctx, userID, locale, description)
}

func (f *fakeStorage) UpdateSocialByAdmin(ctx context.Context, userID, facebook, twitter, instagram string) error {
	if f.UpdateSocialByAdminFn == nil {
		return nil
	}

	return f.UpdateSocialByAdminFn(ctx, userID, facebook, twitter, instagram)
}

func (f *fakeStorage) UpdateSettingsByAdmin(ctx context.Context, userID, locale, timezone string) error {
	if f.UpdateSettingsByAdminFn == nil {
		return nil
	}

	return f.UpdateSettingsByAdminFn(ctx, userID, locale, timezone)
}

func (f *fakeStorage) UpdateNotificationsByAdmin(ctx context.Context, userID string, notifications []string) error {
	if f.UpdateNotificationsByAdminFn == nil {
		return nil
	}

	return f.UpdateNotificationsByAdminFn(ctx, userID, notifications)
}

func (f *fakeStorage) GetTokensForAdmin(ctx context.Context, userID string) (any, error) {
	if f.GetTokensForAdminFn == nil {
		return nil, nil
	}

	return f.GetTokensForAdminFn(ctx, userID)
}

func (f *fakeStorage) DeleteTokenByIDByAdmin(ctx context.Context, tokenID string) error {
	if f.DeleteTokenByIDByAdminFn == nil {
		return nil
	}

	return f.DeleteTokenByIDByAdminFn(ctx, tokenID)
}

func (f *fakeStorage) SetAbilitiesToUser(ctx context.Context, userID string, abilities []string) error {
	if f.SetAbilitiesToUserFn == nil {
		return nil
	}

	return f.SetAbilitiesToUserFn(ctx, userID, abilities)
}

func (f *fakeStorage) AddAbilitiesToUser(ctx context.Context, userID string, abilities []string) error {
	if f.AddAbilitiesToUserFn == nil {
		return nil
	}

	return f.AddAbilitiesToUserFn(ctx, userID, abilities)
}

func (f *fakeStorage) RemoveAbilitiesFromUser(ctx context.Context, userID string, abilities []string) error {
	if f.RemoveAbilitiesFromUserFn == nil {
		return nil
	}

	return f.RemoveAbilitiesFromUserFn(ctx, userID, abilities)
}

// setup initializes the global dependencies of the handlers and resets the injected functions at the end of the test.
func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})
	testAccountFunctions = account.Storage{Tokens: testTokens, Accounts: testAccounts}

	testConfig = config.NewConfiguration()
	testConfig.SupportedLocales = []string{"en"}

	t.Cleanup(func() {
		*testStorage = fakeStorage{}
		*testTokens = fakeAccountTokens{}
		*testAccounts = fakeAccountAccounts{}
		storageRemoved = false
		testAccountFunctions = account.Storage{Tokens: testTokens, Accounts: testAccounts}
	})
}

// serve sends a request with the given JSON body to the handler, mounted on "/:id", and returns the HTTP status.
func serve(handler gin.HandlerFunc, body string) int {
	router := gin.New()
	router.Use(ctxkey.Config.Inject(testConfig), ctxkey.Queue.Inject(testQueue), ctxkey.Timezones.Inject(testTimezones), ctxkey.Notifications.Inject(notifications.NewRegistry()), Inject(testStorage), withoutStorageIf(), account.Inject(&testAccountFunctions), ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
	router.Any("/:id", handler)

	req, _ := http.NewRequest(http.MethodPost, "/42", strings.NewReader(body))

	return apitest.Serve(router, req).Code
}

// newTestContext returns a context carrying the JSON body and the "en" locale.
func newTestContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	c, recorder := apitest.NewContext(http.MethodPost, "/42", strings.NewReader(body))
	ctxkey.Config.Set(c, testConfig)
	ctxkey.Queue.Set(c, testQueue)
	ctxkey.Timezones.Set(c, testTimezones)
	ctxkey.Notifications.Set(c, notifications.NewRegistry())
	SetStorage(c, testStorage)
	account.SetStorage(c, &testAccountFunctions)
	ctxLocale.Set(c, "en", "en")

	return c, recorder
}

// withoutStorageIf is a middleware removing the storage if the test asks for it with withoutStorage.
func withoutStorageIf() gin.HandlerFunc {
	return func(c *gin.Context) {
		if storageRemoved {
			SetStorage(c, nil)
		}
	}
}

// fakeAccountTokens is an account.TokensStorage whose GetTokenDetailsByID calls the function set by the test, if any.
type fakeAccountTokens struct {
	account.TokensStorage
	detailsFn func(ctx context.Context, tokenID string) (authentication.Token, error)
}

func (f *fakeAccountTokens) GetTokenDetailsByID(ctx context.Context, tokenID string) (authentication.Token, error) {
	if f.detailsFn == nil {
		return authentication.Token{}, nil
	}

	return f.detailsFn(ctx, tokenID)
}

// fakeAccountAccounts is an account.AccountsStorage whose methods used by the handlers call the function set by the
// test, if any.
type fakeAccountAccounts struct {
	account.AccountsStorage
	detailsFn func(ctx context.Context, userID string) (authentication.User, error)
	deleteFn  func(ctx context.Context, userID string) error
	disableFn func(ctx context.Context, userID string) (authentication.User, error)
	enableFn  func(ctx context.Context, userID string) (authentication.User, error)
}

func (f *fakeAccountAccounts) GetUserDetailsByID(ctx context.Context, userID string) (authentication.User, error) {
	if f.detailsFn == nil {
		return authentication.User{}, nil
	}

	return f.detailsFn(ctx, userID)
}

func (f *fakeAccountAccounts) DeleteUserByID(ctx context.Context, userID string) error {
	if f.deleteFn == nil {
		return nil
	}

	return f.deleteFn(ctx, userID)
}

func (f *fakeAccountAccounts) DisableUserByID(ctx context.Context, userID string) (authentication.User, error) {
	if f.disableFn == nil {
		return authentication.User{}, nil
	}

	return f.disableFn(ctx, userID)
}

func (f *fakeAccountAccounts) EnableUserByID(ctx context.Context, userID string) (authentication.User, error) {
	if f.enableFn == nil {
		return authentication.User{}, nil
	}

	return f.enableFn(ctx, userID)
}
