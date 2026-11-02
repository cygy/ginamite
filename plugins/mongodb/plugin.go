// Package mongodb is the plugin that stores the data of the application in MongoDB: the users, their auth tokens,
// their notifications and the contacts. It plugs the functions of the API and of the workers into a ginamite.Server.
package mongodb

import (
	"context"
	"strconv"
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/queue"
	apiFunctions "github.com/cygy/ginamite/plugins/mongodb/api/functions"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	workerFunctions "github.com/cygy/ginamite/plugins/mongodb/worker/functions"
	"github.com/cygy/ginamite/worker"

	"github.com/gin-gonic/gin"
)

// Service is the name of MongoDB as the database.service of the configuration.
const Service = "mongodb"

// Plugin is the MongoDB plugin. Create it with NewPlugin, set the properties that differ from the defaults, then
// register it on the server, which calls Configure.
//
// The functions are optional hooks: the application provides them to extend the plugin with its own collections,
// indexes, default documents and documents that reference the users.
type Plugin struct {
	ExtraUserAbilities        []string
	ExtraDefaultUserAbilities []string

	UserMaxValidateRegistrationAttemps uint
	UserMaxDescriptionLength           uint
	UserMaxImageSize                   int64
	UserImagePath                      string
	UserImageWidth                     int64
	UserImageHeight                    int64
	UserImageSmallVersion              bool
	UserImageThumbVersion              bool

	UpdateDocumentsReferencingToDeletedUser               func(ctx context.Context, userID string, mongoSession store.Database) error
	UpdateDocumentsReferencingToDisabledUser              func(ctx context.Context, userID string, mongoSession store.Database) error
	UpdateDocumentsReferencingToEnabledUser               func(ctx context.Context, userID string, mongoSession store.Database) error
	UpdateDocumentsReferencingToUpdatedUserSocialNetworks func(ctx context.Context, userID string, mongoSession store.Database) error
	SanitizeAccounts                                      func(mongoSession store.Database) uint

	CollectionsToRemove    func() []string
	CollectionsToCreate    func() []workerFunctions.CollectionToCreate
	IndexesToRemove        func() []workerFunctions.IndexToDelete
	IndexesToCreate        func() []workerFunctions.IndexToCreate
	CreateDefaultDocuments func(string, string, store.Database)
	RemoveDefaultDocuments func(string, string, store.Database)
	UpdateDefaultDocuments func(string, string, store.Database)
	StartUpTasks           func()

	IsUserEmpty func(context.Context, string, store.Database) bool

	session *session.Session // The session of the database, opened by Connect.
}

// NewPlugin returns a Plugin with the default settings of the model package.
func NewPlugin() *Plugin {
	settings := model.DefaultSettings()

	return &Plugin{
		ExtraUserAbilities:                 []string{},
		ExtraDefaultUserAbilities:          []string{},
		UserMaxValidateRegistrationAttemps: settings.UserMaxValidateRegistrationAttemps,
		UserMaxDescriptionLength:           settings.UserMaxDescriptionLength,
		UserMaxImageSize:                   settings.UserMaxImageSize,
		UserImagePath:                      settings.UserImagePath,
		UserImageWidth:                     settings.UserImageWidth,
		UserImageHeight:                    settings.UserImageHeight,
		UserImageSmallVersion:              settings.UserImageSmallVersion,
		UserImageThumbVersion:              settings.UserImageThumbVersion,
	}
}

var _ ginamite.Connector = (*Plugin)(nil)

// Connect opens the MongoDB session with the database settings of the main configuration. It does nothing if
// the database service of the configuration is not MongoDB. The server calls it before Configure, if it is the API
// or the worker. It returns an error if the server is not reachable.
func (p *Plugin) Connect(configuration *config.Configuration) error {
	if configuration.Database.Service != Service {
		return nil
	}

	database := configuration.Database
	s := session.NewSession()
	if err := s.Initialize(database.Host+":"+strconv.Itoa(database.Port),
		database.Database,
		database.Username,
		database.Password,
		time.Duration(database.TimeOut)); err != nil {
		return err
	}
	p.session = s

	return nil
}

// Close closes the MongoDB session, if it is open.
func (p *Plugin) Close() {
	if p.session != nil {
		p.session.Close()
	}
}

// Configure sets the functions of the plugin on the worker and the API of the server, and the deduplicator of the
// messages, which remembers them in the database. The API saves the settings of
// the plugin, see Settings, in the context of the requests, and adds the routes of the plugin.
func (p *Plugin) Configure(server *ginamite.Server) error {
	p.configureWorker(server.Worker)
	p.configureAPI(server.API)
	server.Deduplicator = p.workers().Deduplicator()

	return nil
}

// Settings returns the settings of the plugin, with the extra abilities added to the default ones.
func (p *Plugin) Settings() model.Settings {
	settings := model.DefaultSettings()
	settings.UserMaxValidateRegistrationAttemps = p.UserMaxValidateRegistrationAttemps
	settings.UserMaxDescriptionLength = p.UserMaxDescriptionLength
	settings.UserMaxImageSize = p.UserMaxImageSize
	settings.UserImagePath = p.UserImagePath
	settings.UserImageWidth = p.UserImageWidth
	settings.UserImageHeight = p.UserImageHeight
	settings.UserImageSmallVersion = p.UserImageSmallVersion
	settings.UserImageThumbVersion = p.UserImageThumbVersion
	settings.AllUserAbilities = append(settings.AllUserAbilities, p.ExtraUserAbilities...)
	settings.DefaultUserAbilities = append(settings.DefaultUserAbilities, p.ExtraDefaultUserAbilities...)

	return settings
}

// database returns the database of the session of the plugin.
func (p *Plugin) database() store.Database {
	return p.session.Database()
}

// workers returns the functions of the workers working on the database of the plugin.
func (p *Plugin) workers() *workerFunctions.Functions {
	f := workerFunctions.New(p.database)
	f.Hooks = workerFunctions.Hooks{
		SanitizeAccounts:                                      p.SanitizeAccounts,
		IsUserEmpty:                                           p.IsUserEmpty,
		UpdateDocumentsReferencingToDeletedUser:               p.UpdateDocumentsReferencingToDeletedUser,
		UpdateDocumentsReferencingToDisabledUser:              p.UpdateDocumentsReferencingToDisabledUser,
		UpdateDocumentsReferencingToEnabledUser:               p.UpdateDocumentsReferencingToEnabledUser,
		UpdateDocumentsReferencingToUpdatedUserSocialNetworks: p.UpdateDocumentsReferencingToUpdatedUserSocialNetworks,
	}

	return f
}

// isValidID reports whether id is the hexadecimal representation of a MongoDB ObjectID.
func isValidID(id string) bool {
	_, err := bson.ObjectIDFromHex(id)
	return err == nil
}

// configureWorker sets the storages and the hooks of the worker.
func (p *Plugin) configureWorker(w *worker.Server) {
	workers := p.workers()

	w.Storage.Tokens = workers
	w.Storage.Accounts = workers
	w.Storage.Users = workers
	w.Storage.Notifications = workers
	w.Storage.Locations = workers
	w.Storage.Emails = workers
	w.Hooks.StartUpTasks = p.startUpTasks
}

// startUpTasks creates and cleans the collections, the indexes and the default documents, when the worker starts.
// The hooks of the plugin run after the default ones.
func (p *Plugin) startUpTasks(ctx context.Context, environment, version string) {
	workers := p.workers()

	workers.RemoveUnusedCollections(ctx)
	if p.CollectionsToRemove != nil {
		collections := p.CollectionsToRemove()
		workers.RemoveCollections(ctx, collections)
	}

	workers.CreateAllCollections(ctx)
	if p.CollectionsToCreate != nil {
		collections := p.CollectionsToCreate()
		workers.CreateCollections(ctx, collections)
	}

	workers.RemoveUnusedIndexes(ctx)
	if p.IndexesToRemove != nil {
		indexes := p.IndexesToRemove()
		workers.RemoveIndexes(ctx, indexes)
	}

	workers.CreateAllIndexes(ctx)
	if p.IndexesToCreate != nil {
		indexes := p.IndexesToCreate()
		workers.CreateIndexes(ctx, indexes)
	}

	workerFunctions.RemoveUnusedDefaultDocuments(environment, version)
	if p.RemoveDefaultDocuments != nil {
		workers.RemoveDefaultDocuments(environment, version, p.RemoveDefaultDocuments)
	}

	workerFunctions.CreateAllDefaultDocuments(environment, version)
	if p.CreateDefaultDocuments != nil {
		workers.CreateDefaultDocuments(environment, version, p.CreateDefaultDocuments)
	}

	workerFunctions.UpdateAllDefaultDocuments(environment, version)
	if p.UpdateDefaultDocuments != nil {
		workers.UpdateDefaultDocuments(environment, version, p.UpdateDefaultDocuments)
	}

	if p.StartUpTasks != nil {
		p.StartUpTasks()
	}
}

// configureAPI sets the storages and the hooks of the API, which work on the database of the plugin, and adds its
// routes.
func (p *Plugin) configureAPI(a *api.Server) {
	a.Hooks.IsValidID = isValidID

	tokens := apiFunctions.NewTokens(p.database)
	a.Hooks.BuildTokenFromID = tokens.BuildTokenFromID
	a.Hooks.ExtendTokenExpirationDateFromID = tokens.ExtendTokenExpirationDateFromID
	a.Account.Tokens = tokens
	a.Account.Processes = apiFunctions.NewProcesses(p.database)
	a.Account.Accounts = apiFunctions.NewAccounts(p.database, p.Settings,
		func(locale string) model.RegistrationDefaults {
			return model.RegistrationDefaults{
				Timezone:      model.RegistrationTimezone(locale, a.Timezones, a.Config.DefaultTimezone),
				Notifications: a.Notifications.Defaults(),
				Abilities:     p.Settings().DefaultUserAbilities,
			}
		},
		func(userID string) { a.Queue.DeleteUser(queue.TaskDeleteUser{UserID: userID}) },
	)

	a.Hooks.MiddlewareIsValidAuthToken = tokens.IsValid
	a.Hooks.MiddlewareGetUserAndLocaleFromAuthToken = apiFunctions.GetUserAndLocaleFromAuthToken

	a.Terms = apiFunctions.NewTerms(p.database)
	a.Account.Notifications = apiFunctions.NewNotifications(p.database, func() string { return a.Config.DefaultLocale })
	users := apiFunctions.NewUsers(p.database, p.Settings)
	a.Users = users
	a.Hooks.MiddlewareGetUserAbilities = users.Abilities

	a.Contacts = apiFunctions.NewContacts(p.database)

	a.AddCustomRoutes(addRoutes(apiFunctions.NewPublicAccount(p.database, p.Settings)))
}

// addRoutes adds the routes of the plugin: the image and the public profile of the account.
func addRoutes(account *apiFunctions.PublicAccount) route.ConfigureCustomRoutes {
	return func(v *gin.RouterGroup, handlers route.Middlewares, _ bool, _ func(admin bool, abilities []string) gin.HandlerFunc) {
		mustBeAuthenticated := handlers.MustBeAuthenticated
		mustAuthenticationBeValid := handlers.MustAuthenticationBeValid
		mustHaveAcceptedTheLatestVersionOfTerms := handlers.MustHaveAcceptedTheLatestVersionOfTerms

		v.PUT("/account/image", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.UpdateImage)
		v.DELETE("/account/image", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.DeleteImage)
		v.PUT("/account/public", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.UpdateProfile)
	}
}
