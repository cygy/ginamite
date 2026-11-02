package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/api/route/terms"
	"github.com/cygy/ginamite/api/route/user"

	"github.com/gin-gonic/gin"
)

// newDefaultRoutes returns the default routes of a router, all enabled.
func newDefaultRoutes() route.DefaultRoutes {
	return route.NewRouter(gin.New(), "v1", false, "secret", "", nil, nil, nil, nil).DefaultRoutes
}

// fill sets every function of the struct pointed by groups to a function doing nothing.
func fill(groups ...any) {
	for _, group := range groups {
		values := reflect.ValueOf(group).Elem()
		for i := range values.NumField() {
			field := values.Field(i)
			field.Set(reflect.MakeFunc(field.Type(), func([]reflect.Value) []reflect.Value {
				out := make([]reflect.Value, field.Type().NumOut())
				for j := range out {
					out[j] = reflect.Zero(field.Type().Out(j))
				}

				return out
			}))
		}
	}
}

// defineAllFunctions makes the server define every function of the features.
func defineAllFunctions(s *Server) {
	fill(&s.Account, &s.Users, &s.Terms)
	s.Contacts = noopContacts{}
}

func TestServer_selectRoutes(t *testing.T) {
	t.Run("no feature", func(t *testing.T) {
		routes := newDefaultRoutes()

		needsJWT, err := NewServer().selectRoutes(&routes, nil)
		if err != nil || needsJWT {
			t.Fatalf("selectRoutes() = %v, %v, want no error and no need of the secret", needsJWT, err)
		}

		if routes.Status.Paths.Get.Enabled || routes.Locales.Paths.Get.Enabled || routes.Timezones.Paths.Get.Enabled {
			t.Error("without feature, no route is loaded")
		}
		if routes.Contact.Paths.Post.Enabled || routes.Admin.Paths.GetContacts.Enabled || routes.Authentication.Paths.LoginWithPassword.Enabled {
			t.Error("the routes of the features which are not enabled must not be loaded")
		}
	})

	t.Run("the features which depend on nothing need no secret", func(t *testing.T) {
		routes := newDefaultRoutes()

		needsJWT, err := NewServer().selectRoutes(&routes, []string{FeatureStatus, FeatureLocales, FeatureTimezones})
		if err != nil || needsJWT {
			t.Fatalf("selectRoutes() = %v, %v, want no error and no need of the secret", needsJWT, err)
		}
		if !routes.Status.Paths.Get.Enabled || !routes.Locales.Paths.Get.Enabled || !routes.Timezones.Paths.Get.Enabled {
			t.Error("the routes of the features must be loaded")
		}
	})

	t.Run("a feature loads its administration routes", func(t *testing.T) {
		s := NewServer()
		defineAllFunctions(s)
		routes := newDefaultRoutes()

		needsJWT, err := s.selectRoutes(&routes, []string{FeatureContact})
		if err != nil || !needsJWT {
			t.Fatalf("selectRoutes() = %v, %v, want no error and the secret needed", needsJWT, err)
		}

		if !routes.Contact.Paths.Post.Enabled || !routes.Admin.Paths.GetContacts.Enabled || !routes.Admin.Paths.UpdateContact.Enabled {
			t.Error("the routes of the contacts and of their administration must be loaded")
		}
		if routes.Admin.Paths.GetEnabledUsers.Enabled || routes.Account.Paths.GetDetails.Enabled {
			t.Error("the routes of the other features must not be loaded")
		}
	})

	t.Run("undefined storages are named", func(t *testing.T) {
		s := NewServer()
		routes := newDefaultRoutes()

		_, err := s.selectRoutes(&routes, []string{FeatureContact})
		if err == nil || !strings.Contains(err.Error(), "Contacts") {
			t.Errorf("error = %v, want the undefined storage", err)
		}
	})

	t.Run("unknown feature", func(t *testing.T) {
		routes := newDefaultRoutes()

		_, err := NewServer().selectRoutes(&routes, []string{"contacts"})
		if err == nil || !strings.Contains(err.Error(), "[contacts]") || !strings.Contains(err.Error(), FeatureContact) {
			t.Errorf("error = %v, want the unknown feature and the known ones", err)
		}
	})
}

func TestFeatures(t *testing.T) {
	names := Features()

	for _, want := range []string{FeatureStatus, FeatureTerms, FeatureAuth, FeatureAccount, FeatureContact} {
		if !slices.Contains(names, want) {
			t.Errorf("Features() = %v, want %q", names, want)
		}
	}
}

// TestServer_features_areConsistent fails when a default route is added without being in a feature, when a route is in
// two features, or when a requirement names a function which is not in a group of functions.
// storages are the names of the storages of the server which the features need.
var storages = []string{"Contacts"}

// noopContacts is a contact.Storage doing nothing.
type noopContacts struct{}

func (noopContacts) CreateContact(context.Context, string, string, string) error { return nil }
func (noopContacts) UpdateContact(context.Context, string, bool) error           { return nil }
func (noopContacts) GetContacts(context.Context, int64, int64) (any, error)      { return nil, nil }
func (noopContacts) GetContact(context.Context, string) (any, error)             { return nil, nil }

// functionGroups are the types of the groups of functions of the server.
var functionGroups = []reflect.Type{reflect.TypeFor[account.Functions](), reflect.TypeFor[user.Functions](), reflect.TypeFor[terms.Functions]()}

func TestServer_features_areConsistent(t *testing.T) {
	routes := newDefaultRoutes()

	listed := map[*route.PathProperties]string{}
	for _, f := range NewServer().features(&routes) {
		for _, r := range f.routes {
			if other, ok := listed[r.route]; ok {
				t.Errorf("the route %q is in the features %s and %s", r.route.Path, other, f.name)
			}
			listed[r.route] = f.name

			for _, function := range r.functions {
				if slices.Contains(storages, function) {
					continue
				}
				if !slices.ContainsFunc(functionGroups, func(group reflect.Type) bool { _, ok := group.FieldByName(function); return ok }) {
					t.Errorf("the route %q needs %q, which is not a function of the server", r.route.Path, function)
				}
			}
		}
	}

	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		if v.Type() == reflect.TypeFor[route.PathProperties]() {
			path := v.Addr().Interface().(*route.PathProperties)
			if _, ok := listed[path]; !ok {
				t.Errorf("the route %q is in no feature: add it to one of them in features", path.Path)
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

func TestRouter_DisableAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	router := route.NewRouter(engine, "v1", false, "", "", nil, nil, nil, nil)
	router.DisableAuthentication()
	engine.GET("/protected", router.Handlers.MustBeAuthenticated, func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/admin", router.Rights(true, nil), func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, target := range []string{"/protected", "/admin"} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: status = %d, want %d: nothing can be authenticated without a secret", target, rec.Code, http.StatusInternalServerError)
		}
	}
}
