package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/ctxnotifications"
	"github.com/cygy/ginamite/common/notifications"

	"github.com/cygy/ginamite/common/ctxtimezone"

	"github.com/cygy/ginamite/common/ctxlocalization"

	"github.com/cygy/ginamite/common/ctxqueue"
	"github.com/cygy/ginamite/common/queue"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
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

	// testFunctions and testAccountFunctions are the functions of the handlers, reset at the end of each test.
	testFunctions        Functions
	testAccountFunctions account.Functions
)

// setup initializes the global dependencies of the handlers and resets the injected functions at the end of the test.
func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	testConfig = config.NewConfiguration()
	testConfig.SupportedLocales = []string{"en"}

	t.Cleanup(func() {
		testFunctions = Functions{}
		testAccountFunctions = account.Functions{}
	})
}

// serve sends a request with the given JSON body to the handler, mounted on "/:id", and returns the HTTP status.
func serve(handler gin.HandlerFunc, body string) int {
	router := gin.New()
	router.Use(ctxconfig.Inject(testConfig), ctxqueue.Inject(testQueue), ctxtimezone.Inject(testTimezones), ctxnotifications.Inject(notifications.NewRegistry()), Inject(&testFunctions), account.Inject(&testAccountFunctions), ctxlocalization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
	router.Any("/:id", handler)

	req, _ := http.NewRequest(http.MethodPost, "/42", strings.NewReader(body))

	return apitest.Serve(router, req).Code
}

// newTestContext returns a context carrying the JSON body and the "en" locale.
func newTestContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	c, recorder := apitest.NewContext(http.MethodPost, "/42", strings.NewReader(body))
	ctxconfig.Set(c, testConfig)
	ctxqueue.Set(c, testQueue)
	ctxtimezone.Set(c, testTimezones)
	ctxnotifications.Set(c, notifications.NewRegistry())
	SetFunctions(c, &testFunctions)
	account.SetFunctions(c, &testAccountFunctions)
	ctxLocale.Set(c, "en", "en")

	return c, recorder
}
