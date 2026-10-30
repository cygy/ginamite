package route

import (
	"net/http"

	"github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/localization"
	apiResponse "github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/response"
	"github.com/cygy/ginamite/web/route/robots"

	"github.com/gin-gonic/gin"
)

// errorPages are the keys of the error pages, in the routes file, and their handlers.
var errorPages = []struct {
	key     string
	handler gin.HandlerFunc
}{
	{"error.internal_server", response.InternalServerError},
	{"error.not_found", response.NotFound},
	{"error.unauthorized", response.Unauthorized},
	{"error.authentication.invalid", response.InvalidAuthentication},
	{"error.authentication.expired", response.ExpiredAuthentication},
	{"error.authentication.revoked", response.RevokedAuthentication},
}

// NewRouter returns a router of the engine for the locales, with the robots.txt route enabled.
// The auth token of the requests is read from the cookie cookieName, and verified with the secret JWTSecret.
func NewRouter(e *gin.Engine, locales []string, defaultLocale, cookieName, JWTSecret string) *Router {
	r := &Router{
		engine:        e,
		locales:       locales,
		defaultLocale: defaultLocale,
		cookieName:    cookieName,
		jwtSecret:     JWTSecret,
	}

	r.Handlers.MustBeAuthenticated = mustBeAuthenticated

	r.DefaultRoutes.Robots.Enabled = true
	r.DefaultRoutes.Robots.Paths.Get.Enabled = true
	r.DefaultRoutes.Robots.Paths.Get.Path = "/robots.txt"

	return r
}

// LoadDefaultRoutes loads the default routes: robots.txt, the redirection of the index to the page of the default
// locale, the page not found for the unknown routes, and, for each locale, the error pages and the async calls.
func (r *Router) LoadDefaultRoutes() {
	if r.DefaultRoutes.Robots.Enabled && r.DefaultRoutes.Robots.Paths.Get.Enabled {
		r.engine.GET(r.DefaultRoutes.Robots.Paths.Get.Path, robots.Robots)
	}

	if path, _, ok := relativeRoutePath("index", r.defaultLocale); ok {
		r.engine.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusPermanentRedirect, "/"+r.defaultLocale+path)
		})
	}

	r.engine.NoRoute(context.SaveUnknownLocale(r.locales, r.defaultLocale), response.NotFound)

	for _, locale := range r.locales {
		group := r.localeGroup(locale)

		for _, page := range errorPages {
			if path, _, ok := relativeRoutePath(page.key, locale); ok {
				group.GET(path, page.handler)
			}
		}

		r.loadAsyncRoutes(group, locale)
	}
}

// LoadCustomRoutes loads the custom routes of the application, for each locale.
func (r *Router) LoadCustomRoutes(configureRoutes ConfigureCustomRoutes) {
	if configureRoutes == nil {
		return
	}

	for _, locale := range r.locales {
		configureRoutes(r.localeGroup(locale), locale, r.Handlers, func(key string) (string, bool, gin.HandlerFunc, bool) {
			path, partial, ok := relativeRoutePath(key, locale)
			return path, partial, context.SavePageKey(key), ok
		})
	}
}

// localeGroup returns the group of the routes of the locale, whose requests have the locale and the auth token saved
// to their context.
func (r *Router) localeGroup(locale string) *gin.RouterGroup {
	return r.engine.Group("/"+locale, context.SaveLocale(locale), context.SaveAuthToken(r.cookieName, r.jwtSecret))
}

// mustBeAuthenticated returns the handler which aborts the requests without auth token: with a JSON error if the
// page is partial, else with the page "error.unauthorized".
func mustBeAuthenticated(partial bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(context.GetAuthTokenID(c)) > 0 {
			return
		}

		if !partial {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		t := localization.Translate(context.GetLocale(c))
		c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse.NewError(apiResponse.Error{
			Domain:   auth.Domain,
			Code:     auth.InsufficientRights,
			Message:  t("error.unauthorized.message"),
			Reason:   t("error.unauthorized.invalid_token.reason"),
			Recovery: t("error.unauthorized.invalid_token.recovery"),
		}))
	}
}
