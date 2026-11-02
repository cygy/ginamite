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
	"github.com/cygy/ginamite/plugins/mongodb/api/ctxsettings"
	apiFunctions "github.com/cygy/ginamite/plugins/mongodb/api/functions"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	workerFunctions "github.com/cygy/ginamite/plugins/mongodb/worker/functions"
	"github.com/cygy/ginamite/worker"

	"github.com/gin-gonic/gin"
)

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
	if !configuration.IsMongoDBDatabase() {
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
	return workerFunctions.New(p.database)
}

// isValidID reports whether id is the hexadecimal representation of a MongoDB ObjectID.
func isValidID(id string) bool {
	_, err := bson.ObjectIDFromHex(id)
	return err == nil
}

// configureWorker sets the functions of the worker.
func (p *Plugin) configureWorker(w *worker.Server) {
	workers := p.workers()

	w.Functions.DeleteExpiredTokens = workers.DeleteExpiredTokens
	w.Functions.GetNeverUsedAccounts = workers.GetNeverUsedAccounts
	w.Functions.GetInactiveAccounts = workers.GetInactiveAccounts
	w.Functions.IsUserEmpty = workers.IsEmptyUser(p.IsUserEmpty)
	w.Functions.SanitizeAccounts = workers.SanitizeAccounts(p.SanitizeAccounts)
	w.Functions.DeleteUserByID = workers.DeleteUserByID(p.UpdateDocumentsReferencingToDeletedUser)
	w.Functions.DisableUserByID = workers.DisableUserByID(p.UpdateDocumentsReferencingToDisabledUser)
	w.Functions.EnableUserByID = workers.EnableUserByID(p.UpdateDocumentsReferencingToEnabledUser)
	w.Functions.UpdateUserSocialNetworksByID = workers.UpdateUserSocialNetworksByID(p.UpdateDocumentsReferencingToUpdatedUserSocialNetworks)
	w.Functions.SaveIPAddressDetails = workers.SaveIPAddressDetails
	w.Functions.GetNotificationTargetsByUserAndType = workers.GetNotificationTargetsByUserAndType
	w.Functions.IsEmailAddressValid = workers.IsEmailAddressValid
	w.Functions.StartUpTasks = p.startUpTasks
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

// configureAPI sets the functions of the API, the validation of the identifiers and the middleware which saves the
// database in the context of the requests, if the session is open, and adds its routes.
func (p *Plugin) configureAPI(a *api.Server) {
	a.Functions.IsValidID = isValidID
	a.Middlewares = append(a.Middlewares, ctxsettings.InjectSettings(p.Settings()))
	if p.session != nil {
		a.Middlewares = append(a.Middlewares, p.session.Inject())
	}

	tokens := apiFunctions.NewTokens(p.database)
	a.Functions.BuildTokenFromID = tokens.BuildTokenFromID
	a.Functions.ExtendTokenExpirationDateFromID = tokens.ExtendTokenExpirationDateFromID

	a.Functions.MiddlewareIsValidAuthToken = apiFunctions.IsValidAuthToken
	a.Functions.MiddlewareGetUserAndLocaleFromAuthToken = apiFunctions.GetUserAndLocaleFromAuthToken
	a.Functions.MiddlewareGetUserAbilities = apiFunctions.GetUserAbilities
	a.Functions.MiddlewareGetLatestVersionOfTermsAcceptedByUser = apiFunctions.GetLatestVersionOfTermsAcceptedByUser

	a.Functions.GetEnabledUsersForAdmin = apiFunctions.GetEnabledUsersForAdmin
	a.Functions.GetDisabledUsersForAdmin = apiFunctions.GetDisabledUsersForAdmin
	a.Functions.UpdatePasswordByAdmin = apiFunctions.UpdateUserPasswordByAdmin
	a.Functions.UpdateInformationsByAdmin = apiFunctions.UpdateUserInformationsByAdmin
	a.Functions.UpdateDescriptionByAdmin = apiFunctions.UpdateUserDescriptionByAdmin
	a.Functions.UpdateSocialByAdmin = apiFunctions.UpdateUserSocialByAdmin
	a.Functions.UpdateSettingsByAdmin = apiFunctions.UpdateUserSettingsByAdmin
	a.Functions.UpdateNotificationsByAdmin = apiFunctions.UpdateUserNotificationsByAdmin
	a.Functions.GetTokensForAdmin = apiFunctions.GetTokensForAdmin
	a.Functions.DeleteTokenByIDByAdmin = apiFunctions.DeleteTokenByIDByAdmin
	a.Functions.SetAbilitiesToUser = apiFunctions.SetAbilitiesToUser
	a.Functions.AddAbilitiesToUser = apiFunctions.AddAbilitiesToUser
	a.Functions.RemoveAbilitiesFromUser = apiFunctions.RemoveAbilitiesFromUser

	a.Functions.NotificationsReceivedByUser = apiFunctions.NotificationsReceivedByUser
	a.Functions.NotificationReceivedByUser = apiFunctions.NotificationReceivedByUser
	a.Functions.UpdateUserNotificationRead = apiFunctions.UpdateUserNotificationRead
	a.Functions.UpdateUserNotificationNotified = apiFunctions.UpdateUserNotificationNotified
	a.Functions.DeleteUserNotification = apiFunctions.DeleteUserNotification
	a.Functions.NotificationsSubscribedByUser = apiFunctions.NotificationsSubscribedByUser
	a.Functions.SubscribeNotificationsToUser = apiFunctions.SubscribeNotificationsToUser
	a.Functions.UnsubscribeNotificationFromUser = apiFunctions.UnsubscribeNotificationFromUser

	a.Functions.GetUserDetailsByIdentifier = apiFunctions.GetUserDetailsByIdentifier
	a.Functions.GetUserDetailsByID = apiFunctions.GetUserDetailsByID
	a.Functions.GetUserDetailsByFacebookUserID = apiFunctions.GetUserDetailsByFacebookUserID
	a.Functions.GetUserDetailsByGoogleUserID = apiFunctions.GetUserDetailsByGoogleUserID
	a.Functions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser = apiFunctions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser

	a.Functions.UpdateEncryptedPasswordByUserID = apiFunctions.UpdateEncryptedPasswordByUserID
	a.Functions.NewForgotPasswordProcessForUser = apiFunctions.NewForgotPasswordProcessForUser
	a.Functions.GetForgotPasswordProcessDetailsByID = apiFunctions.GetForgotPasswordProcessDetailsByID
	a.Functions.WrongAccessToForgotPasswordProcess = apiFunctions.WrongAccessToForgotPasswordProcess
	a.Functions.DeleteForgotPasswordProcess = apiFunctions.DeleteForgotPasswordProcess

	a.Functions.UpdateUserPassword = apiFunctions.UpdateUserPassword
	a.Functions.UpdateUserEmailAddress = apiFunctions.UpdateUserEmailAddress
	a.Functions.UserEmailAddressIsVerified = apiFunctions.UserEmailAddressIsVerified
	a.Functions.UpdateUserSocialNetworks = apiFunctions.UpdateUserSocialNetworks
	a.Functions.UpdateVersionOfTermsAcceptedByUser = apiFunctions.UpdateVersionOfTermsAcceptedByUser

	a.Functions.ValidateUserRegistration = apiFunctions.ValidateUserRegistration
	a.Functions.CancelUserRegistration = apiFunctions.CancelUserRegistration
	a.Functions.GetRegistrationDetails = apiFunctions.GetRegistrationDetails
	a.Functions.RegisterByEmailAddress = apiFunctions.RegisterByEmailAddress
	a.Functions.RegisterByThirdPartyToken = apiFunctions.RegisterByThirdPartyToken

	a.Functions.DoesAccountWithUsernameExist = apiFunctions.DoesAccountWithUsernameExist
	a.Functions.DoesAccountWithEmailAddressExist = apiFunctions.DoesAccountWithEmailAddressExist
	a.Functions.DoesUserWithFacebookUserIDExist = apiFunctions.DoesUserWithFacebookUserIDExist
	a.Functions.DoesUserWithGoogleUserIDExist = apiFunctions.DoesUserWithGoogleUserIDExist

	a.Functions.SaveFacebookUserTokenDetailsToUser = apiFunctions.SaveFacebookUserTokenDetailsToUser
	a.Functions.SaveGoogleUserTokenDetailsToUser = apiFunctions.SaveGoogleUserTokenDetailsToUser

	a.Functions.SaveAuthenticationToken = apiFunctions.SaveAuthenticationToken
	a.Functions.GetTokenExpirationDate = apiFunctions.GetTokenExpirationDate
	a.Functions.DeleteTokenByID = apiFunctions.DeleteTokenByID
	a.Functions.GetTokenDetailsByID = apiFunctions.GetTokenDetailsByID
	a.Functions.GetOwnedTokens = apiFunctions.GetOwnedTokens
	a.Functions.GetOwnedTokenByID = apiFunctions.GetOwnedTokenByID
	a.Functions.UpdateTokenByID = apiFunctions.UpdateTokenByID

	a.Functions.GetOwnedAccountDetails = apiFunctions.GetOwnedAccountDetails
	a.Functions.UpdatedAccountInfos = apiFunctions.UpdatedAccountInfos
	a.Functions.UpdatedAccountSettings = apiFunctions.UpdatedAccountSettings

	a.Functions.CreateNewDeleteAccountProcess = apiFunctions.CreateNewDeleteAccountProcess
	a.Functions.CreateNewDisableAccountProcess = apiFunctions.CreateNewDisableAccountProcess
	a.Functions.CreateNewUpdatePasswordProcess = apiFunctions.CreateNewUpdatePasswordProcess
	a.Functions.CreateNewUpdateEmailAddressProcess = apiFunctions.CreateNewUpdateEmailAddressProcess
	a.Functions.CreateNewVerifyEmailAddressProcess = apiFunctions.CreateNewVerifyEmailAddressProcess
	a.Functions.VerifyProcess = apiFunctions.VerifyProcess
	a.Functions.DeleteProcessByID = apiFunctions.DeleteProcessByID

	a.Functions.DeleteUserByID = apiFunctions.DeleteUserByID
	a.Functions.DisableUserByID = apiFunctions.DisableUserByID
	a.Functions.EnableUserByID = apiFunctions.EnableUserByID

	a.Functions.CreateContact = apiFunctions.CreateContact
	a.Functions.UpdateContact = apiFunctions.UpdateContact
	a.Functions.GetContacts = apiFunctions.GetContactsForAdmin
	a.Functions.GetContact = apiFunctions.GetContactForAdmin

	a.AddCustomRoutes(addRoutes)
}

// addRoutes adds the routes of the plugin: the image and the public profile of the account.
func addRoutes(v *gin.RouterGroup, handlers route.Middlewares, isProductionEnvironment bool, rights func(admin bool, abilities []string) gin.HandlerFunc) {
	mustBeAuthenticated := handlers.MustBeAuthenticated
	mustAuthenticationBeValid := handlers.MustAuthenticationBeValid
	mustHaveAcceptedTheLatestVersionOfTerms := handlers.MustHaveAcceptedTheLatestVersionOfTerms

	v.PUT("/account/image", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, apiFunctions.UpdateAccountPublicImage)
	v.DELETE("/account/image", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, apiFunctions.DeleteAccountPublicImage)
	v.PUT("/account/public", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, apiFunctions.UpdateAccountPublicProfile)
}
