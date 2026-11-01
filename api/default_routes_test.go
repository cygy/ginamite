package api

import (
	"reflect"
	"testing"

	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// newDefaultRoutes returns the default routes of a router, all enabled.
func newDefaultRoutes() route.DefaultRoutes {
	return route.NewRouter(gin.New(), "v1", false, "secret", "", nil, nil, nil, nil).DefaultRoutes
}

func TestServer_disableUndefinedRoutes(t *testing.T) {
	t.Run("nothing defined", func(t *testing.T) {
		routes := newDefaultRoutes()
		NewServer().disableUndefinedRoutes(&routes)

		for name, path := range map[string]route.PathProperties{
			"status":   routes.Status.Paths.Get,
			"locales":  routes.Locales.Paths.Get,
			"terms":    routes.Terms.Paths.GetTerms,
			"timezone": routes.Timezones.Paths.Get,
		} {
			if !path.Enabled {
				t.Errorf("%s: the route without function must stay enabled", name)
			}
		}

		for name, path := range map[string]route.PathProperties{
			"login":    routes.Authentication.Paths.LoginWithPassword,
			"register": routes.Registration.Paths.RegisterWithPassword,
			"contact":  routes.Contact.Paths.Post,
			"admin":    routes.Admin.Paths.GetContact,
			"account":  routes.Account.Paths.GetDetails,
		} {
			if path.Enabled {
				t.Errorf("%s: the route whose functions are undefined must be disabled", name)
			}
		}
	})

	t.Run("some functions defined", func(t *testing.T) {
		s := NewServer()
		s.Functions.CreateContact = func(*gin.Context, string, string, string) error { return nil }
		routes := newDefaultRoutes()
		s.disableUndefinedRoutes(&routes)

		if !routes.Contact.Paths.Post.Enabled {
			t.Error("the route Post of the contacts must stay enabled")
		}
		if routes.Admin.Paths.UpdateContact.Enabled {
			t.Error("the route UpdateContact must be disabled")
		}
	})

	t.Run("warns when the hashes can not be upgraded", func(t *testing.T) {
		s := NewServer()
		s.Functions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser = func(*gin.Context, string) (authentication.User, string, error) {
			return authentication.User{}, "", nil
		}
		s.Functions.SaveAuthenticationToken = func(*gin.Context, authentication.Details) (string, bool, error) { return "", false, nil }

		routes := newDefaultRoutes()
		if got := countWarnings(func() { s.disableUndefinedRoutes(&routes) }); got != 1 {
			t.Errorf("warnings = %d, want 1", got)
		}

		routes = newDefaultRoutes()
		if got := countWarnings(func() { NewServer().disableUndefinedRoutes(&routes) }); got != 0 {
			t.Errorf("warnings = %d, want 0 when the login is disabled", got)
		}
	})

	t.Run("a route needing two functions", func(t *testing.T) {
		s := NewServer()
		s.Functions.GetTokenDetailsByID = func(*gin.Context, string) (authentication.Token, error) { return authentication.Token{}, nil }
		routes := newDefaultRoutes()
		s.disableUndefinedRoutes(&routes)

		if routes.Authentication.Paths.UpdateToken.Enabled {
			t.Error("UpdateToken needs UpdateTokenByID too")
		}
	})
}

// TestServer_requirements_coverEveryRoute fails when a default route is added without its requirements: either it is
// listed, or it is one of the routes which call no function of the application.
func TestServer_requirements_coverEveryRoute(t *testing.T) {
	routes := newDefaultRoutes()

	listed := map[*route.PathProperties]bool{}
	for _, r := range NewServer().requirements(&routes) {
		listed[r.route] = true
	}

	alwaysAvailable := map[*route.PathProperties]bool{
		&routes.Status.Paths.Get:        true,
		&routes.Locales.Paths.Get:       true,
		&routes.Timezones.Paths.Get:     true,
		&routes.Notifications.Paths.Get: true,
		&routes.Terms.Paths.GetTerms:    true,
		&routes.Terms.Paths.GetVersion:  true,
	}

	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		if v.Type() == reflect.TypeFor[route.PathProperties]() {
			path := v.Addr().Interface().(*route.PathProperties)
			if !listed[path] && !alwaysAvailable[path] {
				t.Errorf("the route %q has no requirement: list it in requirements, or among the routes always available", path.Path)
			}

			return
		}

		if v.Kind() == reflect.Struct {
			for i := range v.NumField() {
				walk(v.Field(i))
			}
		}
	}
	walk(reflect.ValueOf(&routes).Elem())
}
