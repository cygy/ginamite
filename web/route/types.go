package route

import "github.com/gin-gonic/gin"

// ConfigureDefaultRoutes : function to configure the default routes of the web server.
type ConfigureDefaultRoutes func(routes *DefaultRoutes)

// ConfigureCustomRoutes : function to configure the custom routes of the web server.
type ConfigureCustomRoutes func(g *gin.RouterGroup, locale string, handlers Middlewares, getRelativePath func(key string) (string, bool, gin.HandlerFunc, bool))

// Router : properties to define routes and their rights.
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

// Middlewares : collection of the built-in handlers.
type Middlewares struct {
	MustBeAuthenticated func(bool) gin.HandlerFunc
}

// PathProperties : properties of a route.
type PathProperties struct {
	Enabled bool
	Path    string
}

// DefaultRoutes : definition of the built-in routes.
type DefaultRoutes struct {
	Robots struct {
		Enabled bool
		Paths   struct {
			Get PathProperties
		}
	}
}
