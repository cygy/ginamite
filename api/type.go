package api

import (
	tokencache "github.com/cygy/ginamite/api/cache/authtoken"
	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	"github.com/cygy/ginamite/api/request"
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

	// The functions and the storages of the application, such as the ones of the storage plugin, by feature. A feature
	// whose functions are undefined can not be enabled, see api_features.
	Account  account.Storage
	Users    user.Storage
	Contacts contact.Storage
	Terms    terms.Storage
	Hooks    Hooks

	// RevokedTokens are the auth tokens revoked, which are rejected until they expire. The cache messages of the queue
	// add tokens to it.
	RevokedTokens *tokencache.Revoked

	// Middlewares are added to the Gin engine, in order, before the middleware that sets the locale and the routes.
	// They are meant for the needs of the application, such as a request ID or a CORS policy.
	Middlewares []gin.HandlerFunc

	customRoutes []route.ConfigureCustomRoutes
}

// Hooks are the functions of the application which are not tied to a feature: the validation of the identifiers and of
// the auth tokens, the tokens signed by the API, and the extensions of the server. A database plugin sets most of them.
type Hooks struct {
	// RecurringTasks is run once when the API starts, to start the periodic tasks of the application. Optional.
	RecurringTasks func()
	// IsValidID reports whether a string is a valid identifier of the storage, such as in the paths of the routes. The
	// identifiers are not checked if nil.
	IsValidID func(id string) bool
	// NewMessageFromQueueCache handles the cache messages of the queue the API does not know, and reports whether it
	// handled the message. Optional.
	NewMessageFromQueueCache func(message string) bool
	// MiddlewareIsValidAuthToken reports whether the auth token of a request is still valid: not expired nor revoked.
	MiddlewareIsValidAuthToken ctxAuthToken.IsValidFunc
	// MiddlewareGetUserAndLocaleFromAuthToken loads the user and the locale of an auth token.
	MiddlewareGetUserAndLocaleFromAuthToken ctxAuthToken.GetUserAndLocaleFunc
	// MiddlewareGetUserAbilities returns the abilities of a user and whether the user is an administrator.
	MiddlewareGetUserAbilities rights.GetUserAbilitiesFunc
	// BuildTokenFromID builds the auth token stored under an identifier.
	BuildTokenFromID authentication.BuildTokenFunc
	// ExtraPropertiesForTokenWithID returns the extra properties signed in the auth token of an identifier. Optional.
	ExtraPropertiesForTokenWithID authentication.ExtraPropertiesForTokenWithIDFunc
	// ExtendTokenExpirationDateFromID extends the expiration date of the auth token of an identifier.
	ExtendTokenExpirationDateFromID authentication.ExtendTokenExpirationDateFunc
	// RouterConfigureDefaultRoutes changes the default routes before they are registered, to replace or remove some of
	// them. Optional.
	RouterConfigureDefaultRoutes route.ConfigureDefaultRoutes

	// NewRateLimiter returns the limiters of the requests which try to guess a secret, see the setting rate_limits. The
	// default limiters keep their counters in memory, by instance of the API.
	NewRateLimiter request.NewRateLimiterFunc
}
