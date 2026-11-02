package mongodb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"testing"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/rights"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/config"
	apiFunctions "github.com/cygy/ginamite/plugins/mongodb/api/functions"
	_ "github.com/cygy/ginamite/plugins/mongodb/internal/plugintest" // Initializes the logger and gin.
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
	workerFunctions "github.com/cygy/ginamite/plugins/mongodb/worker/functions"

	"github.com/gin-gonic/gin"
)

func TestNewPlugin(t *testing.T) {
	p := NewPlugin()

	if got, want := p.Settings(), model.DefaultSettings(); !reflect.DeepEqual(got, want) {
		t.Errorf("the plugin does not start with the default settings: %+v, want %+v", got, want)
	}
	if p.ExtraUserAbilities == nil || p.ExtraDefaultUserAbilities == nil {
		t.Error("the abilities must be empty lists")
	}
}

// newConfiguration returns a configuration whose database service is the one given.
func newConfiguration(service string) *config.Configuration {
	configuration := config.NewConfiguration()
	configuration.Database.Service = service

	return configuration
}

func TestPlugin_Connect_OtherService(t *testing.T) {
	p := NewPlugin()

	if err := p.Connect(newConfiguration("postgresql")); err != nil {
		t.Fatal(err)
	}

	if p.session != nil {
		t.Error("the plugin must not connect to MongoDB when the service is another one")
	}
}

func TestPlugin_Connect_Unreachable(t *testing.T) {
	p := NewPlugin()
	configuration := newConfiguration(Service)
	configuration.Database.Host = "127.0.0.1"
	configuration.Database.Port = 1
	configuration.Database.TimeOut = 1

	if err := p.Connect(configuration); err == nil || p.session != nil {
		t.Errorf("Connect() = %v, session = %+v, want an error and no session", err, p.session)
	}
}

func TestPlugin_Close(t *testing.T) {
	t.Run("without session", func(t *testing.T) {
		NewPlugin().Close()
	})

	t.Run("with a session which is not connected", func(t *testing.T) {
		p := NewPlugin()
		p.session = session.NewSession()

		p.Close()
	})
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"507f1f77bcf86cd799439011", true},
		{"", false},
		{"xyz", false},
		{"507f1f77bcf86cd79943901", false},
		{"507f1f77bcf86cd79943901g", false},
		{"123e4567-e89b-12d3-a456-426614174000", false},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := isValidID(tt.id); got != tt.want {
				t.Errorf("isValidID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestPlugin_Configure_IsValidID(t *testing.T) {
	server := ginamite.NewServer()
	if err := NewPlugin().Configure(server); err != nil {
		t.Fatal(err)
	}

	if isValid := server.API.Hooks.IsValidID; isValid == nil || !isValid("507f1f77bcf86cd799439011") || isValid("xyz") {
		t.Error("the API must validate the IDs as ObjectIDs")
	}
}

func TestPlugin_Configure_Deduplicator(t *testing.T) {
	server := ginamite.NewServer()
	memory := server.Deduplicator
	if err := NewPlugin().Configure(server); err != nil {
		t.Fatal(err)
	}

	if server.Deduplicator == nil || server.Deduplicator == memory {
		t.Error("the messages must be remembered in the database")
	}
}

func TestPlugin_Configure_NoMiddleware(t *testing.T) {
	// The functions of the plugin get the database and the settings when they are created: there is nothing to put in
	// the context of the requests.
	server := ginamite.NewServer()
	if err := NewPlugin().Configure(server); err != nil {
		t.Fatal(err)
	}

	if len(server.API.Middlewares) != 0 {
		t.Errorf("len(Middlewares) = %d, want none", len(server.API.Middlewares))
	}
}

func TestPlugin_Settings(t *testing.T) {
	p := NewPlugin()
	p.UserMaxValidateRegistrationAttemps = 9
	p.UserMaxDescriptionLength = 10
	p.UserMaxImageSize = 11
	p.UserImagePath = "/custom"
	p.UserImageWidth = 12
	p.UserImageHeight = 13
	p.UserImageSmallVersion = false
	p.UserImageThumbVersion = false
	p.ExtraUserAbilities = []string{"extra"}
	p.ExtraDefaultUserAbilities = []string{"default"}

	got := p.Settings()

	if got.UserMaxValidateRegistrationAttemps != 9 || got.UserMaxDescriptionLength != 10 || got.UserMaxImageSize != 11 || got.UserImagePath != "/custom" ||
		got.UserImageWidth != 12 || got.UserImageHeight != 13 || got.UserImageSmallVersion || got.UserImageThumbVersion {
		t.Errorf("settings = %+v", got)
	}
	if !slices.Contains(got.AllUserAbilities, "extra") || !slices.Contains(got.AllUserAbilities, rights.AbilityToManageUsers) || !slices.Equal(got.DefaultUserAbilities, []string{"default"}) {
		t.Errorf("abilities = %v, %v", got.AllUserAbilities, got.DefaultUserAbilities)
	}

	// The settings do not share the abilities of the defaults.
	if slices.Contains(model.DefaultSettings().AllUserAbilities, "extra") {
		t.Error("the default settings must not change")
	}
}

func TestPlugin_Configure_Functions(t *testing.T) {
	server := ginamite.NewServer()
	if err := NewPlugin().Configure(server); err != nil {
		t.Fatal(err)
	}

	api, worker := server.API, server.Worker

	if _, ok := api.Account.Accounts.(*apiFunctions.Accounts); !ok {
		t.Errorf("Accounts = %T, want the storage of the plugin", api.Account.Accounts)
	}
	if _, ok := api.Contacts.(*apiFunctions.Contacts); !ok {
		t.Errorf("Contacts = %T, want the storage of the plugin", api.Contacts)
	}

	// The functions working on the database, and the storages of the worker, are defined too.
	for name, defined := range map[string]bool{
		"MiddlewareIsValidAuthToken":      api.Hooks.MiddlewareIsValidAuthToken != nil,
		"MiddlewareGetUserAbilities":      api.Hooks.MiddlewareGetUserAbilities != nil,
		"BuildTokenFromID":                api.Hooks.BuildTokenFromID != nil,
		"ExtendTokenExpirationDateFromID": api.Hooks.ExtendTokenExpirationDateFromID != nil,
		"Tokens":                          worker.Storage.Tokens != nil,
		"Accounts":                        worker.Storage.Accounts != nil,
		"Users":                           worker.Storage.Users != nil,
		"Notifications":                   worker.Storage.Notifications != nil,
		"Locations":                       worker.Storage.Locations != nil,
		"Emails":                          worker.Storage.Emails != nil,
		"StartUpTasks":                    worker.Hooks.StartUpTasks != nil,
	} {
		if !defined {
			t.Errorf("worker.%s is not defined", name)
		}
	}
}

func TestPlugin_Configure_WorkerHooks(t *testing.T) {
	db := mock.NewDatabase("app")

	var sanitized, deleted, empty bool
	p := NewPlugin()
	p.session = &session.Session{Store: db}
	p.SanitizeAccounts = func(store.Database) uint { sanitized = true; return 5 }
	p.UpdateDocumentsReferencingToDeletedUser = func(context.Context, string, store.Database) error { deleted = true; return nil }
	p.IsUserEmpty = func(context.Context, string, store.Database) bool { empty = true; return false }

	server := ginamite.NewServer()
	if err := p.Configure(server); err != nil {
		t.Fatal(err)
	}

	if got, err := server.Worker.Storage.Accounts.Sanitize(context.Background()); err != nil || got != 5 || !sanitized {
		t.Errorf("Sanitize() = %d, %v", got, err)
	}
	// The ID is not an ObjectID: the deletion fails, after the hook ran.
	if _, err := server.Worker.Storage.Users.Delete(context.Background(), "user-1"); err == nil {
		t.Error("an invalid user ID must fail")
	}
	if !deleted {
		t.Error("the hook that updates the documents of a deleted user is not called")
	}
	if got, err := server.Worker.Storage.Accounts.IsEmpty(context.Background(), "user-1"); err != nil || got || !empty {
		t.Error("IsEmpty() does not use the hook")
	}
}

func TestPlugin_startUpTasks(t *testing.T) {
	db := mock.NewDatabase("app")

	t.Run("without hooks", func(t *testing.T) {
		p := NewPlugin()
		p.session = &session.Session{Store: db}
		p.startUpTasks(context.Background(), "test", "1.0")

		for _, name := range []string{model.UserCollection, model.AuthTokenCollection} {
			if !slices.Contains(db.Created(), name) {
				t.Errorf("the collection %s must be created", name)
			}
			if view := db.Coll(name).IndexView; view == nil || len(view.Created) == 0 {
				t.Errorf("the indexes of %s must be created", name)
			}
		}
	})

	t.Run("hooks run after the default tasks, in order", func(t *testing.T) {
		db = mock.NewDatabase("app")

		var steps []string
		step := func(name string) { steps = append(steps, name) }
		documents := func(name string) func(string, string, store.Database) {
			return func(environment, version string, _ store.Database) {
				if environment != "test" || version != "1.0" {
					t.Errorf("%s: arguments = %q, %q", name, environment, version)
				}
				step(name)
			}
		}

		p := NewPlugin()
		p.session = &session.Session{Store: db}
		p.CollectionsToRemove = func() []string { step("remove collections"); return []string{"old"} }
		p.CollectionsToCreate = func() []workerFunctions.CollectionToCreate {
			step("create collections")
			return []workerFunctions.CollectionToCreate{{Name: "custom"}}
		}
		p.IndexesToRemove = func() []workerFunctions.IndexToDelete {
			step("remove indexes")
			return []workerFunctions.IndexToDelete{{Collection: "custom", IndexName: "a_1"}}
		}
		p.IndexesToCreate = func() []workerFunctions.IndexToCreate {
			step("create indexes")
			return []workerFunctions.IndexToCreate{{Collection: "custom"}}
		}
		p.RemoveDefaultDocuments = documents("remove documents")
		p.CreateDefaultDocuments = documents("create documents")
		p.UpdateDefaultDocuments = documents("update documents")
		p.StartUpTasks = func() { step("start up") }

		p.startUpTasks(context.Background(), "test", "1.0")

		want := []string{"remove collections", "create collections", "remove indexes", "create indexes", "remove documents", "create documents", "update documents", "start up"}
		if !slices.Equal(steps, want) {
			t.Errorf("steps = %v, want %v", steps, want)
		}

		if !slices.Contains(db.Created(), "custom") || len(db.Coll("old").Calls()) == 0 {
			t.Error("the custom collections must be created and removed")
		}
		if view := db.Coll("custom").IndexView; view == nil || len(view.Created) != 1 || len(view.Dropped) != 1 {
			t.Errorf("custom indexes = %+v", view)
		}
	})
}

func TestAddRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	var order []string
	middleware := func(name string) gin.HandlerFunc {
		return func(c *gin.Context) { order = append(order, name) }
	}
	// The last middleware stops the request: the handlers need a user and a database.
	last := func(c *gin.Context) {
		order = append(order, "terms")
		c.AbortWithStatus(http.StatusForbidden)
	}
	addRoutes(apiFunctions.NewPublicAccount(func() store.Database { return mock.NewDatabase("app") }, model.DefaultSettings))(engine.Group("/v1"), route.Middlewares{
		MustBeAuthenticated:                     middleware("authenticated"),
		MustAuthenticationBeValid:               middleware("valid"),
		MustHaveAcceptedTheLatestVersionOfTerms: last,
	}, true, nil)

	got := map[string]bool{}
	for _, r := range engine.Routes() {
		got[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{"PUT /v1/account/image", "DELETE /v1/account/image", "PUT /v1/account/public"} {
		if !got[want] {
			t.Errorf("the route %s is missing, routes: %v", want, got)
		}
	}

	// The routes are protected: the middlewares run in order, before the handler.
	for _, route := range []struct{ method, path string }{{http.MethodPut, "/v1/account/image"}, {http.MethodDelete, "/v1/account/image"}, {http.MethodPut, "/v1/account/public"}} {
		order = nil
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(route.method, route.path, nil))

		if !slices.Equal(order, []string{"authenticated", "valid", "terms"}) || recorder.Code != http.StatusForbidden {
			t.Errorf("%s %s: middlewares = %v, status = %d", route.method, route.path, order, recorder.Code)
		}
	}
}
