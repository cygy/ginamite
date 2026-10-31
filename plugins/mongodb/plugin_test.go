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
	"github.com/cygy/ginamite/common/config/database"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/api/ctxsettings"
	apiFunctions "github.com/cygy/ginamite/plugins/mongodb/api/functions"
	_ "github.com/cygy/ginamite/plugins/mongodb/internal/plugintest" // Initializes the logger and gin.
	"github.com/cygy/ginamite/plugins/mongodb/model"
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

	p.Connect(newConfiguration("postgresql"))

	if p.session != nil {
		t.Error("the plugin must not connect to MongoDB when the service is another one")
	}
}

func TestPlugin_Connect_Unreachable(t *testing.T) {
	p := NewPlugin()
	configuration := newConfiguration(database.MongoDB)
	configuration.Database.Host = "127.0.0.1"
	configuration.Database.Port = 1
	configuration.Database.TimeOut = 1

	defer func() {
		if recover() == nil {
			t.Error("Connect() must panic when the server is unreachable")
		}
		if p.session == nil || p.session.DatabaseName != configuration.Database.Database {
			t.Errorf("the session must be created with the database of the configuration: %+v", p.session)
		}
	}()

	p.Connect(configuration)
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
	NewPlugin().Configure(server)

	if isValid := server.API.Functions.IsValidID; isValid == nil || !isValid("507f1f77bcf86cd799439011") || isValid("xyz") {
		t.Error("the API must validate the IDs as ObjectIDs")
	}
}

func TestPlugin_Configure_Middlewares(t *testing.T) {
	t.Run("without session", func(t *testing.T) {
		server := ginamite.NewServer()
		NewPlugin().Configure(server)

		if len(server.API.Middlewares) != 1 {
			t.Errorf("len(Middlewares) = %d, want only the one of the settings", len(server.API.Middlewares))
		}
	})

	t.Run("with a session", func(t *testing.T) {
		db := mock.NewDatabase("app")
		p := NewPlugin()
		p.session = &session.Session{Store: db}

		server := ginamite.NewServer()
		p.Configure(server)

		if len(server.API.Middlewares) != 2 {
			t.Fatalf("len(Middlewares) = %d, want 2", len(server.API.Middlewares))
		}

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		server.API.Middlewares[1](c)

		if session.Get(c) != store.Database(db) {
			t.Error("the middleware must save the database in the context")
		}
	})
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

func TestPlugin_Configure_Settings(t *testing.T) {
	p := NewPlugin()
	p.UserMaxDescriptionLength = 10

	server := ginamite.NewServer()
	p.Configure(server)

	if len(server.API.Middlewares) == 0 {
		t.Fatal("the middleware of the settings is missing")
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	server.API.Middlewares[0](c)

	if got := ctxsettings.GetSettings(c).UserMaxDescriptionLength; got != 10 {
		t.Errorf("UserMaxDescriptionLength = %d, want 10", got)
	}
}

// sameFunction reports whether the two values are the same function.
func sameFunction(a, b any) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

func TestPlugin_Configure_Functions(t *testing.T) {
	server := ginamite.NewServer()
	NewPlugin().Configure(server)

	api, worker := server.API.Functions, server.Worker.Functions

	// The functions which do not need the database are set as they are.
	for name, pair := range map[string][2]any{
		"MiddlewareIsValidAuthToken": {api.MiddlewareIsValidAuthToken, apiFunctions.IsValidAuthToken},
		"GetUserDetailsByID":         {api.GetUserDetailsByID, apiFunctions.GetUserDetailsByID},
		"RegisterByEmailAddress":     {api.RegisterByEmailAddress, apiFunctions.RegisterByEmailAddress},
		"CreateContact":              {api.CreateContact, apiFunctions.CreateContact},
		"GetContacts":                {api.GetContacts, apiFunctions.GetContactsForAdmin},
		"GetOwnedAccountDetails":     {api.GetOwnedAccountDetails, apiFunctions.GetOwnedAccountDetails},
	} {
		if pair[0] == nil || !sameFunction(pair[0], pair[1]) {
			t.Errorf("%s is not the function of the plugin", name)
		}
	}

	// The functions working on the database, and the ones built from the hooks of the plugin, are defined too.
	for name, defined := range map[string]bool{
		"BuildTokenFromID":                    api.BuildTokenFromID != nil,
		"ExtendTokenExpirationDateFromID":     api.ExtendTokenExpirationDateFromID != nil,
		"DeleteExpiredTokens":                 worker.DeleteExpiredTokens != nil,
		"GetNeverUsedAccounts":                worker.GetNeverUsedAccounts != nil,
		"GetInactiveAccounts":                 worker.GetInactiveAccounts != nil,
		"SaveIPAddressDetails":                worker.SaveIPAddressDetails != nil,
		"IsEmailAddressValid":                 worker.IsEmailAddressValid != nil,
		"GetNotificationTargetsByUserAndType": worker.GetNotificationTargetsByUserAndType != nil,
		"IsUserEmpty":                         worker.IsUserEmpty != nil,
		"SanitizeAccounts":                    worker.SanitizeAccounts != nil,
		"DeleteUserByID":                      worker.DeleteUserByID != nil,
		"DisableUserByID":                     worker.DisableUserByID != nil,
		"EnableUserByID":                      worker.EnableUserByID != nil,
		"UpdateUserSocialNetworksByID":        worker.UpdateUserSocialNetworksByID != nil,
		"StartUpTasks":                        worker.StartUpTasks != nil,
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
	p.UpdateDocumentsReferencingToDeletedUser = func(context.Context, string, store.Database) { deleted = true }
	p.IsUserEmpty = func(context.Context, string, store.Database) bool { empty = true; return false }

	server := ginamite.NewServer()
	p.Configure(server)

	if got := server.Worker.Functions.SanitizeAccounts(); got != 5 || !sanitized {
		t.Errorf("SanitizeAccounts() = %d", got)
	}
	server.Worker.Functions.DeleteUserByID(context.Background(), "user-1")
	if !deleted {
		t.Error("the hook that updates the documents of a deleted user is not called")
	}
	if server.Worker.Functions.IsUserEmpty(context.Background(), "user-1") || !empty {
		t.Error("IsUserEmpty() does not use the hook")
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
	addRoutes(engine.Group("/v1"), route.Middlewares{
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
