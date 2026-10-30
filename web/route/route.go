// Package route sets up the routes of the web server: the built-in ones (robots.txt, error pages, passthrough
// asynchronous calls to the API) and the custom ones of the application, for each supported locale.
package route

import "github.com/gin-gonic/gin"

// ConfigureDefaultRoutes is the function configuring the default routes of the web server.
type ConfigureDefaultRoutes func(routes *DefaultRoutes)

// ConfigureCustomRoutes is the function configuring the custom routes of the web server for a locale.
// The group is the one of the locale, and getRelativePath returns, for the key of a route of the routes file, its
// path relative to the group, whether the page is partial, the handler saving its key to the context, and whether
// the route exists for the locale.
type ConfigureCustomRoutes func(g *gin.RouterGroup, locale string, handlers Middlewares, getRelativePath func(key string) (string, bool, gin.HandlerFunc, bool))

// Router holds the properties to define the routes and their rights.
type Router struct {
	DefaultRoutes DefaultRoutes
	Handlers      Middlewares

	// private properties
	engine        *gin.Engine
	locales       []string
	defaultLocale string
	cookieName    string
	jwtSecret     string
}

// Middlewares is the collection of the built-in handlers.
type Middlewares struct {
	// MustBeAuthenticated returns the handler which rejects the requests without auth token. The responses of a partial
	// page are JSON ones.
	MustBeAuthenticated func(partial bool) gin.HandlerFunc
}

// PathProperties holds the properties of a route.
type PathProperties struct {
	Enabled bool
	Path    string
}

// DefaultRoutes is the definition of the built-in routes.
type DefaultRoutes struct {
	Robots struct {
		Enabled bool
		Paths   struct {
			Get PathProperties
		}
	}
}
