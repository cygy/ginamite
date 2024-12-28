package mongodb

import (
	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/route"
	apiFunctions "github.com/cygy/ginamite/plugins/mongodb/api/functions"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	workerFunctions "github.com/cygy/ginamite/plugins/mongodb/worker/functions"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

// Plugin : structure of the plugin.
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

	UpdateDocumentsReferencingToDeletedUser               func(userID string, mongoSession *mgo.Database)
	UpdateDocumentsReferencingToDisabledUser              func(userID string, mongoSession *mgo.Database)
	UpdateDocumentsReferencingToEnabledUser               func(userID string, mongoSession *mgo.Database)
	UpdateDocumentsReferencingToUpdatedUserSocialNetworks func(userID string, mongoSession *mgo.Database)
	UpdateDocumentsReferencingToDeletedUsers              func(mongoSession *mgo.Database) uint

	CollectionsToRemove    func() []string
	CollectionsToCreate    func() []workerFunctions.CollectionToCreate
	IndexesToRemove        func() []workerFunctions.IndexToDelete
	IndexesToCreate        func() []workerFunctions.IndexToCreate
	CreateDefaultDocuments func(string, string, *mgo.Database)
	RemoveDefaultDocuments func(string, string, *mgo.Database)
	UpdateDefaultDocuments func(string, string, *mgo.Database)
	StartUpTasks           func()

	IsUserEmpty func(string, *mgo.Database) bool
}

// NewPlugin : returns a new struct 'Plugin'.
func NewPlugin() *Plugin {
	return &Plugin{
		ExtraUserAbilities:                 []string{},
		ExtraDefaultUserAbilities:          []string{},
		UserMaxValidateRegistrationAttemps: model.UserMaxValidateRegistrationAttemps,
		UserMaxDescriptionLength:           model.UserMaxDescriptionLength,
		UserMaxImageSize:                   model.UserMaxImageSize,
		UserImagePath:                      model.UserImagePath,
		UserImageWidth:                     model.UserImageWidth,
		UserImageHeight:                    model.UserImageHeight,
		UserImageSmallVersion:              model.UserImageSmallVersion,
		UserImageThumbVersion:              model.UserImageThumbVersion,
	}
}

// Configure : configure the plugin.
func (p *Plugin) Configure(server *ginamite.Server) {
	model.UserMaxValidateRegistrationAttemps = p.UserMaxValidateRegistrationAttemps
	model.UserMaxDescriptionLength = p.UserMaxDescriptionLength
	model.UserMaxImageSize = p.UserMaxImageSize
	model.UserImagePath = p.UserImagePath
	model.UserImageWidth = p.UserImageWidth
	model.UserImageHeight = p.UserImageHeight
	model.UserImageSmallVersion = p.UserImageSmallVersion
	model.UserImageThumbVersion = p.UserImageThumbVersion

	model.AddUserAbilities(p.ExtraUserAbilities)
	model.AddDefaultUserAbilities(p.ExtraDefaultUserAbilities)

	server.Worker.Functions.DeleteExpiredTokens = workerFunctions.DeleteExpiredTokens
	server.Worker.Functions.GetNeverUsedAccounts = workerFunctions.GetNeverUsedAccounts
	server.Worker.Functions.GetInactiveAccounts = workerFunctions.GetInactiveAccounts
	server.Worker.Functions.IsUserEmpty = workerFunctions.IsEmptyUser(p.IsUserEmpty)
	server.Worker.Functions.CleanDeletedAccounts = workerFunctions.CleanDeletedAccounts(p.UpdateDocumentsReferencingToDeletedUsers)
	server.Worker.Functions.DeleteUserByID = workerFunctions.DeleteUserByID(p.UpdateDocumentsReferencingToDeletedUser)
	server.Worker.Functions.DisableUserByID = workerFunctions.DisableUserByID(p.UpdateDocumentsReferencingToDisabledUser)
	server.Worker.Functions.EnableUserByID = workerFunctions.EnableUserByID(p.UpdateDocumentsReferencingToEnabledUser)
	server.Worker.Functions.UpdateUserSocialNetworksByID = workerFunctions.UpdateUserSocialNetworksByID(p.UpdateDocumentsReferencingToUpdatedUserSocialNetworks)
	server.Worker.Functions.SaveIPAddressDetails = workerFunctions.SaveIPAddressDetails
	server.Worker.Functions.GetNotificationTargetsByUserAndType = workerFunctions.GetNotificationTargetsByUserAndType
	server.Worker.Functions.IsEmailAddressValid = workerFunctions.IsEmailAddressValid
	server.Worker.Functions.StartUpTasks = func(environment, version string) {
		workerFunctions.RemoveUnusedCollections()
		if p.CollectionsToRemove != nil {
			collections := p.CollectionsToRemove()
			workerFunctions.RemoveCollections(collections)
		}

		workerFunctions.CreateAllCollections()
		if p.CollectionsToCreate != nil {
			collections := p.CollectionsToCreate()
			workerFunctions.CreateCollections(collections)
		}

		workerFunctions.RemoveUnusedIndexes()
		if p.IndexesToRemove != nil {
			indexes := p.IndexesToRemove()
			workerFunctions.RemoveIndexes(indexes)
		}

		workerFunctions.CreateAllIndexes()
		if p.IndexesToCreate != nil {
			indexes := p.IndexesToCreate()
			workerFunctions.CreateIndexes(indexes)
		}

		workerFunctions.RemoveUnusedDefaultDocuments(environment, version)
		if p.RemoveDefaultDocuments != nil {
			workerFunctions.RemoveDefaultDocuments(environment, version, p.RemoveDefaultDocuments)
		}

		workerFunctions.CreateAllDefaultDocuments(environment, version)
		if p.CreateDefaultDocuments != nil {
			workerFunctions.CreateDefaultDocuments(environment, version, p.CreateDefaultDocuments)
		}

		workerFunctions.UpdateAllDefaultDocuments(environment, version)
		if p.UpdateDefaultDocuments != nil {
			workerFunctions.UpdateDefaultDocuments(environment, version, p.UpdateDefaultDocuments)
		}

		if p.StartUpTasks != nil {
			p.StartUpTasks()
		}
	}

	server.API.Functions.BuildTokenFromID = apiFunctions.BuildTokenFromID
	server.API.Functions.ExtendTokenExpirationDateFromID = apiFunctions.ExtendTokenExpirationDateFromID

	server.API.Functions.MiddlewareIsValidAuthToken = apiFunctions.IsValidAuthToken
	server.API.Functions.MiddlewareGetUserAndLocaleFromAuthToken = apiFunctions.GetUserAndLocaleFromAuthToken
	server.API.Functions.MiddlewareGetUserAbilities = apiFunctions.GetUserAbilities
	server.API.Functions.MiddlewareGetLatestVersionOfTermsAcceptedByUser = apiFunctions.GetLatestVersionOfTermsAcceptedByUser

	server.API.Functions.GetEnabledUsersForAdmin = apiFunctions.GetEnabledUsersForAdmin
	server.API.Functions.GetDisabledUsersForAdmin = apiFunctions.GetDisabledUsersForAdmin
	server.API.Functions.UpdatePasswordByAdmin = apiFunctions.UpdateUserPasswordByAdmin
	server.API.Functions.UpdateInformationsByAdmin = apiFunctions.UpdateUserInformationsByAdmin
	server.API.Functions.UpdateDescriptionByAdmin = apiFunctions.UpdateUserDescriptionByAdmin
	server.API.Functions.UpdateSocialByAdmin = apiFunctions.UpdateUserSocialByAdmin
	server.API.Functions.UpdateSettingsByAdmin = apiFunctions.UpdateUserSettingsByAdmin
	server.API.Functions.UpdateNotificationsByAdmin = apiFunctions.UpdateUserNotificationsByAdmin
	server.API.Functions.GetTokensForAdmin = apiFunctions.GetTokensForAdmin
	server.API.Functions.DeleteTokenByIDByAdmin = apiFunctions.DeleteTokenByIDByAdmin
	server.API.Functions.SetAbilitiesToUser = apiFunctions.SetAbilitiesToUser
	server.API.Functions.AddAbilitiesToUser = apiFunctions.AddAbilitiesToUser
	server.API.Functions.RemoveAbilitiesFromUser = apiFunctions.RemoveAbilitiesFromUser

	server.API.Functions.NotificationsReceivedByUser = apiFunctions.NotificationsReceivedByUser
	server.API.Functions.NotificationReceivedByUser = apiFunctions.NotificationReceivedByUser
	server.API.Functions.UpdateUserNotificationRead = apiFunctions.UpdateUserNotificationRead
	server.API.Functions.UpdateUserNotificationNotified = apiFunctions.UpdateUserNotificationNotified
	server.API.Functions.DeleteUserNotification = apiFunctions.DeleteUserNotification
	server.API.Functions.NotificationsSubscribedByUser = apiFunctions.NotificationsSubscribedByUser
	server.API.Functions.SubscribeNotificationsToUser = apiFunctions.SubscribeNotificationsToUser
	server.API.Functions.UnsubscribeNotificationFromUser = apiFunctions.UnsubscribeNotificationFromUser

	server.API.Functions.GetUserDetailsByIdentifier = apiFunctions.GetUserDetailsByIdentifier
	server.API.Functions.GetUserDetailsByID = apiFunctions.GetUserDetailsByID
	server.API.Functions.GetUserDetailsByFacebookUserID = apiFunctions.GetUserDetailsByFacebookUserID
	server.API.Functions.GetUserDetailsByGoogleUserID = apiFunctions.GetUserDetailsByGoogleUserID
	server.API.Functions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser = apiFunctions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser

	server.API.Functions.NewForgotPasswordProcessForUser = apiFunctions.NewForgotPasswordProcessForUser
	server.API.Functions.GetForgotPasswordProcessDetailsByID = apiFunctions.GetForgotPasswordProcessDetailsByID
	server.API.Functions.WrongAccessToForgotPasswordProcess = apiFunctions.WrongAccessToForgotPasswordProcess
	server.API.Functions.DeleteForgotPasswordProcess = apiFunctions.DeleteForgotPasswordProcess

	server.API.Functions.UpdateUserPassword = apiFunctions.UpdateUserPassword
	server.API.Functions.UpdateUserEmailAddress = apiFunctions.UpdateUserEmailAddress
	server.API.Functions.UserEmailAddressIsVerified = apiFunctions.UserEmailAddressIsVerified
	server.API.Functions.UpdateUserSocialNetworks = apiFunctions.UpdateUserSocialNetworks
	server.API.Functions.UpdateVersionOfTermsAcceptedByUser = apiFunctions.UpdateVersionOfTermsAcceptedByUser

	server.API.Functions.ValidateUserRegistration = apiFunctions.ValidateUserRegistration
	server.API.Functions.CancelUserRegistration = apiFunctions.CancelUserRegistration
	server.API.Functions.GetRegistrationDetails = apiFunctions.GetRegistrationDetails
	server.API.Functions.RegisterByEmailAddress = apiFunctions.RegisterByEmailAddress
	server.API.Functions.RegisterByThirdPartyToken = apiFunctions.RegisterByThirdPartyToken

	server.API.Functions.DoesAccountWithUsernameExist = apiFunctions.DoesAccountWithUsernameExist
	server.API.Functions.DoesAccountWithEmailAddressExist = apiFunctions.DoesAccountWithEmailAddressExist
	server.API.Functions.DoesUserWithFacebookUserIDExist = apiFunctions.DoesUserWithFacebookUserIDExist
	server.API.Functions.DoesUserWithGoogleUserIDExist = apiFunctions.DoesUserWithGoogleUserIDExist

	server.API.Functions.SaveFacebookUserTokenDetailsToUser = apiFunctions.SaveFacebookUserTokenDetailsToUser
	server.API.Functions.SaveGoogleUserTokenDetailsToUser = apiFunctions.SaveGoogleUserTokenDetailsToUser

	server.API.Functions.SaveAuthenticationToken = apiFunctions.SaveAuthenticationToken
	server.API.Functions.GetTokenExpirationDate = apiFunctions.GetTokenExpirationDate
	server.API.Functions.DeleteTokenByID = apiFunctions.DeleteTokenByID
	server.API.Functions.GetTokenDetailsByID = apiFunctions.GetTokenDetailsByID
	server.API.Functions.GetOwnedTokens = apiFunctions.GetOwnedTokens
	server.API.Functions.GetOwnedTokenByID = apiFunctions.GetOwnedTokenByID
	server.API.Functions.UpdateTokenByID = apiFunctions.UpdateTokenByID

	server.API.Functions.GetOwnedAccountDetails = apiFunctions.GetOwnedAccountDetails
	server.API.Functions.UpdatedAccountInfos = apiFunctions.UpdatedAccountInfos
	server.API.Functions.UpdatedAccountSettings = apiFunctions.UpdatedAccountSettings

	server.API.Functions.CreateNewDeleteAccountProcess = apiFunctions.CreateNewDeleteAccountProcess
	server.API.Functions.CreateNewDisableAccountProcess = apiFunctions.CreateNewDisableAccountProcess
	server.API.Functions.CreateNewUpdatePasswordProcess = apiFunctions.CreateNewUpdatePasswordProcess
	server.API.Functions.CreateNewUpdateEmailAddressProcess = apiFunctions.CreateNewUpdateEmailAddressProcess
	server.API.Functions.CreateNewVerifyEmailAddressProcess = apiFunctions.CreateNewVerifyEmailAddressProcess
	server.API.Functions.VerifyProcess = apiFunctions.VerifyProcess
	server.API.Functions.DeleteProcessByID = apiFunctions.DeleteProcessByID

	server.API.Functions.DeleteUserByID = apiFunctions.DeleteUserByID
	server.API.Functions.DisableUserByID = apiFunctions.DisableUserByID
	server.API.Functions.EnableUserByID = apiFunctions.EnableUserByID

	server.API.Functions.CreateContact = apiFunctions.CreateContact
	server.API.Functions.UpdateContact = apiFunctions.UpdateContact
	server.API.Functions.GetContacts = apiFunctions.GetContactsForAdmin
	server.API.Functions.GetContact = apiFunctions.GetContactForAdmin

	server.API.AddCustomRoutes(func(v *gin.RouterGroup, handlers route.Middlewares, isProductionEnvironment bool, rights func(admin bool, abilities []string) gin.HandlerFunc) {
		mustBeAuthenticated := handlers.MustBeAuthenticated
		mustAuthenticationBeValid := handlers.MustAuthenticationBeValid
		mustHaveAcceptedTheLatestVersionOfTerms := handlers.MustHaveAcceptedTheLatestVersionOfTerms

		v.PUT("/account/image", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, apiFunctions.UpdateAccountPublicImage)
		v.DELETE("/account/image", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, apiFunctions.DeleteAccountPublicImage)
		v.PUT("/account/public", mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, apiFunctions.UpdateAccountPublicProfile)
	})
}
