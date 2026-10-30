package mongodb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"testing"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	apiFunctions "github.com/cygy/ginamite/plugins/mongodb/api/functions"
	_ "github.com/cygy/ginamite/plugins/mongodb/internal/plugintest" // Initializes the logger and gin.
	"github.com/cygy/ginamite/plugins/mongodb/model"
	workerFunctions "github.com/cygy/ginamite/plugins/mongodb/worker/functions"

	"github.com/gin-gonic/gin"
)

// saveModelSettings restores the settings of the model package at the end of the test.
func saveModelSettings(t *testing.T) {
	t.Helper()

	attempts, length, size, path := model.UserMaxValidateRegistrationAttemps, model.UserMaxDescriptionLength, model.UserMaxImageSize, model.UserImagePath
	width, height, small, thumb := model.UserImageWidth, model.UserImageHeight, model.UserImageSmallVersion, model.UserImageThumbVersion
	all, defaults := slices.Clone(model.AllUserAbilities), slices.Clone(model.DefaultUserAbilities)

	t.Cleanup(func() {
		model.UserMaxValidateRegistrationAttemps, model.UserMaxDescriptionLength, model.UserMaxImageSize, model.UserImagePath = attempts, length, size, path
		model.UserImageWidth, model.UserImageHeight, model.UserImageSmallVersion, model.UserImageThumbVersion = width, height, small, thumb
		model.AllUserAbilities, model.DefaultUserAbilities = all, defaults
	})
}

func TestNewPlugin(t *testing.T) {
	p := NewPlugin()

	if p.UserMaxValidateRegistrationAttemps != model.UserMaxValidateRegistrationAttemps || p.UserMaxDescriptionLength != model.UserMaxDescriptionLength ||
		p.UserMaxImageSize != model.UserMaxImageSize || p.UserImagePath != model.UserImagePath || p.UserImageWidth != model.UserImageWidth ||
		p.UserImageHeight != model.UserImageHeight || p.UserImageSmallVersion != model.UserImageSmallVersion || p.UserImageThumbVersion != model.UserImageThumbVersion {
		t.Errorf("the plugin does not start with the settings of the model: %+v", p)
	}
	if p.ExtraUserAbilities == nil || p.ExtraDefaultUserAbilities == nil {
		t.Error("the abilities must be empty lists")
	}
}

func TestPlugin_Configure_Settings(t *testing.T) {
	saveModelSettings(t)

	p := NewPlugin()
	p.UserMaxValidateRegistrationAttemps = 9
	p.UserMaxDescriptionLength = 10
	p.UserMaxImageSize = 11
	p.UserImagePath = "/custom"
	p.UserImageWidth = 12
	p.UserImageHeight = 13
	p.UserImageSmallVersion = !model.UserImageSmallVersion
	p.UserImageThumbVersion = !model.UserImageThumbVersion
	p.ExtraUserAbilities = []string{"extra"}
	p.ExtraDefaultUserAbilities = []string{"default"}

	p.Configure(ginamite.NewServer())

	if model.UserMaxValidateRegistrationAttemps != 9 || model.UserMaxDescriptionLength != 10 || model.UserMaxImageSize != 11 || model.UserImagePath != "/custom" ||
		model.UserImageWidth != 12 || model.UserImageHeight != 13 || model.UserImageSmallVersion != p.UserImageSmallVersion || model.UserImageThumbVersion != p.UserImageThumbVersion {
		t.Error("the settings of the plugin must be applied to the model")
	}
	if !slices.Contains(model.AllUserAbilities, "extra") || !slices.Contains(model.DefaultUserAbilities, "default") {
		t.Errorf("abilities = %v, %v", model.AllUserAbilities, model.DefaultUserAbilities)
	}
}

// sameFunction reports whether the two values are the same function.
func sameFunction(a, b any) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

func TestPlugin_Configure_Functions(t *testing.T) {
	saveModelSettings(t)

	server := ginamite.NewServer()
	NewPlugin().Configure(server)

	api, worker := server.API.Functions, server.Worker.Functions

	// The functions without parameters of the plugin are set as they are.
	for name, pair := range map[string][2]any{
		"BuildTokenFromID":                    {api.BuildTokenFromID, apiFunctions.BuildTokenFromID},
		"MiddlewareIsValidAuthToken":          {api.MiddlewareIsValidAuthToken, apiFunctions.IsValidAuthToken},
		"GetUserDetailsByID":                  {api.GetUserDetailsByID, apiFunctions.GetUserDetailsByID},
		"RegisterByEmailAddress":              {api.RegisterByEmailAddress, apiFunctions.RegisterByEmailAddress},
		"CreateContact":                       {api.CreateContact, apiFunctions.CreateContact},
		"GetContacts":                         {api.GetContacts, apiFunctions.GetContactsForAdmin},
		"GetOwnedAccountDetails":              {api.GetOwnedAccountDetails, apiFunctions.GetOwnedAccountDetails},
		"worker.DeleteExpiredTokens":          {worker.DeleteExpiredTokens, workerFunctions.DeleteExpiredTokens},
		"worker.GetNeverUsedAccounts":         {worker.GetNeverUsedAccounts, workerFunctions.GetNeverUsedAccounts},
		"worker.SaveIPAddressDetails":         {worker.SaveIPAddressDetails, workerFunctions.SaveIPAddressDetails},
		"worker.IsEmailAddressValid":          {worker.IsEmailAddressValid, workerFunctions.IsEmailAddressValid},
		"worker.GetNotificationTargetsByUser": {worker.GetNotificationTargetsByUserAndType, workerFunctions.GetNotificationTargetsByUserAndType},
	} {
		if pair[0] == nil || !sameFunction(pair[0], pair[1]) {
			t.Errorf("%s is not the function of the plugin", name)
		}
	}

	// The functions built from the hooks of the plugin are defined too.
	for name, defined := range map[string]bool{
		"IsUserEmpty":                  worker.IsUserEmpty != nil,
		"SanitizeAccounts":             worker.SanitizeAccounts != nil,
		"DeleteUserByID":               worker.DeleteUserByID != nil,
		"DisableUserByID":              worker.DisableUserByID != nil,
		"EnableUserByID":               worker.EnableUserByID != nil,
		"UpdateUserSocialNetworksByID": worker.UpdateUserSocialNetworksByID != nil,
		"StartUpTasks":                 worker.StartUpTasks != nil,
	} {
		if !defined {
			t.Errorf("worker.%s is not defined", name)
		}
	}
}

func TestPlugin_Configure_WorkerHooks(t *testing.T) {
	saveModelSettings(t)

	db := mock.NewDatabase("app")
	old := session.Main
	t.Cleanup(func() { session.Main = old })
	session.Main = &session.Session{Store: db}

	var sanitized, deleted, empty bool
	p := NewPlugin()
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
	old := session.Main
	t.Cleanup(func() { session.Main = old })
	session.Main = &session.Session{Store: db}

	t.Run("without hooks", func(t *testing.T) {
		NewPlugin().startUpTasks(context.Background(), "test", "1.0")

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
		session.Main = &session.Session{Store: db}

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
