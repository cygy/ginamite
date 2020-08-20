package route

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/localization"
	apiResponse "github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/response"
	"github.com/cygy/ginamite/web/route/robots"

	"github.com/gin-gonic/gin"
)

// NewRouter : returns a new Router struct.
func NewRouter(e *gin.Engine, locales []string, defaultLocale, cookieName, JWTSecret string) *Router {
	r := &Router{}

	r.engine = e
	r.locales = locales
	r.defaultLocale = defaultLocale
	r.cookieName = cookieName
	r.jwtSecret = JWTSecret

	r.Handlers.MustBeAuthenticated = func(partial bool) gin.HandlerFunc {
		return func(c *gin.Context) {
			authTokenID := context.GetAuthTokenID(c)

			if len(authTokenID) == 0 {

				if partial {
					locale := context.GetLocale(c)
					t := localization.Translate(locale)
					c.AbortWithStatusJSON(http.StatusUnauthorized, apiResponse.NewError(apiResponse.Error{
						Domain:   auth.Domain,
						Code:     auth.InsufficientRights,
						Message:  t("error.unauthorized.message"),
						Reason:   t("error.unauthorized.invalid_token.reason"),
						Recovery: t("error.unauthorized.invalid_token.recovery"),
					}))
				} else {
					html.Main.PageByKey(c, "error.unauthorized").RenderUnauthorized(c)
					c.Abort()
				}
			}
		}
	}

	r.DefaultRoutes.Robots.Enabled = true
	r.DefaultRoutes.Robots.Paths.Get.Enabled = true
	r.DefaultRoutes.Robots.Paths.Get.Path = "/robots.txt"

	return r
}

// LoadDefaultRoutes : load the default routes.
func (r *Router) LoadDefaultRoutes() {
	// ----- robots.txt
	if r.DefaultRoutes.Robots.Enabled {
		if r.DefaultRoutes.Robots.Paths.Get.Enabled {
			r.engine.GET(r.DefaultRoutes.Robots.Paths.Get.Path, robots.Robots)
		}
	}

	// ----- index.
	if path, _, ok := relativeRoutePath("index", r.defaultLocale); ok {
		r.engine.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusPermanentRedirect, "/"+r.defaultLocale+path)
		})
	}

	// The default 404 page.
	r.engine.NoRoute(context.SaveUnknownLocale(r.locales, r.defaultLocale), response.NotFound)

	for _, locale := range r.locales {
		group := r.engine.Group("/"+locale, context.SaveLocale(locale), context.SaveAuthToken(r.cookieName, r.jwtSecret))
		{
			// Load the error pages.
			if path, _, ok := relativeRoutePath("error.internal_server", locale); ok {
				group.GET(path, response.InternalServerError)
			}
			if path, _, ok := relativeRoutePath("error.not_found", locale); ok {
				group.GET(path, response.NotFound)
			}
			if path, _, ok := relativeRoutePath("error.unauthorized", locale); ok {
				group.GET(path, response.Unauthorized)
			}
			if path, _, ok := relativeRoutePath("error.authentication.invalid", locale); ok {
				group.GET(path, response.InvalidAuthentication)
			}
			if path, _, ok := relativeRoutePath("error.authentication.expired", locale); ok {
				group.GET(path, response.ExpiredAuthentication)
			}
			if path, _, ok := relativeRoutePath("error.authentication.revoked", locale); ok {
				group.GET(path, response.RevokedAuthentication)
			}

			// Set up the async calls through API.
			for key, async := range html.Main.Async {
				if path, ok := relativeAsyncPath(key, locale); ok {
					// Some vars are used here to avoid 'f' to capture the struct 'async'
					// and to use the latest 'async' properties for all the calls.
					target := async.Target
					method := async.Method

					f := func(c *gin.Context) {
						fullEndpoint := target
						for _, param := range c.Params {
							fullEndpoint = strings.Replace(fullEndpoint, ":"+param.Key, param.Value, 1)
						}
						endpointParts := strings.Split(c.Request.URL.RequestURI(), "?")
						if len(endpointParts) > 1 {
							fullEndpoint = fmt.Sprintf("%s?%s", fullEndpoint, endpointParts[1])
						}

						body, _ := ioutil.ReadAll(c.Request.Body)
						var res interface{}

						request := api.NewRequest()
						request.Endpoint = fullEndpoint
						request.Method = strings.ToUpper(method)
						request.Body = body
						request.Response = &res
						request.Authenticated = true
						request.UseDefaultResponseHandlers = false
						request.ContentType = c.Request.Header.Get("Content-Type")

						result, _ := api.Main.Call(c, request)
						if result.RequestError != nil {
							c.AbortWithStatus(http.StatusInternalServerError)
							return
						}

						if result.APIError.IsNil() {
							c.AbortWithStatusJSON(result.StatusCode, res)
						} else {
							c.AbortWithStatusJSON(result.StatusCode, struct {
								Error apiResponse.Error `json:"error"`
							}{
								Error: result.APIError,
							})
						}
					}

					switch strings.ToUpper(async.Method) {
					case api.GET:
						group.GET(path, f)
						break
					case api.POST:
						group.POST(path, f)
						break
					case api.PUT:
						group.PUT(path, f)
						break
					case api.DELETE:
						group.DELETE(path, f)
						break
					case api.PATCH:
						group.PATCH(path, f)
						break
					}
				}
			}
		}
	}
}

// LoadCustomRoutes : load the custom routes.
func (r *Router) LoadCustomRoutes(configureRoutes ConfigureCustomRoutes) {
	if configureRoutes == nil {
		return
	}

	for _, locale := range r.locales {
		group := r.engine.Group("/"+locale, context.SaveLocale(locale), context.SaveAuthToken(r.cookieName, r.jwtSecret))
		{
			configureRoutes(group, locale, r.Handlers, func(key string) (string, bool, gin.HandlerFunc, bool) {
				path, partial, ok := relativeRoutePath(key, locale)
				return path, partial, context.SavePageKey(key), ok
			})
		}
	}
}
