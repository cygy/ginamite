package api

import (
	tokencache "github.com/cygy/ginamite/api/cache/authtoken"
	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	"github.com/cygy/ginamite/api/rights"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/api/route/contact"
	"github.com/cygy/ginamite/api/route/terms"
	"github.com/cygy/ginamite/api/route/user"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/metrics"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

// Server defines an API server.
type Server struct {
	// Config is the configuration of the application, set by the main server before it starts.
	Config *config.Configuration
	// Queue sends the messages to the other servers, set by the main server before it starts.
	Queue *queue.Queue
	// Localization translates the messages of the responses, set by the main server before it starts.
	Localization *localization.Engine
	// Timezones are the timezones of the countries, set by the main server before it starts.
	Timezones *timezone.Configuration
	// Notifications is the registry of the notifications the users can subscribe to, set by the main server before it
	// starts.
	Notifications *notifications.Registry
	// Requests counts the requests handled by the API for the metrics. Nothing is counted if nil.
	Requests *metrics.Requests

	// The functions of the application, such as the ones of the storage plugin, by feature. A feature whose functions
	// are undefined can not be enabled, see api_features.
	Account  account.Functions
	Users    user.Functions
	Contacts contact.Functions
	Terms    terms.Functions
	Hooks    Hooks

	// RevokedTokens are the auth tokens revoked, which are rejected until they expire. The cache messages of the queue
	// add tokens to it.
	RevokedTokens *tokencache.Revoked

	// Middlewares are added to the Gin engine, in order, before the middleware that sets the locale and the routes.
	// A database plugin uses them to put its connection in the context of the requests.
	Middlewares []gin.HandlerFunc

	customRoutes []route.ConfigureCustomRoutes
}

// Hooks are the functions of the application which are not tied to a feature: the validation of the identifiers and of
// the auth tokens, the tokens signed by the API, and the extensions of the server. A database plugin sets most of them.
type Hooks struct {
	RecurringTasks                                  func()
	IsValidID                                       func(id string) bool
	NewMessageFromQueueCache                        func(message string) bool
	MiddlewareIsValidAuthToken                      ctxAuthToken.IsValidFunc
	MiddlewareGetUserAndLocaleFromAuthToken         ctxAuthToken.GetUserAndLocaleFunc
	MiddlewareGetUserAbilities                      rights.GetUserAbilitiesFunc
	MiddlewareGetLatestVersionOfTermsAcceptedByUser terms.GetLatestVersionOfTermsAcceptedByUserFunc
	BuildTokenFromID                                authentication.BuildTokenFunc
	ExtraPropertiesForTokenWithID                   authentication.ExtraPropertiesForTokenWithIDFunc
	ExtendTokenExpirationDateFromID                 authentication.ExtendTokenExpirationDateFunc
	RouterConfigureDefaultRoutes                    route.ConfigureDefaultRoutes
}
