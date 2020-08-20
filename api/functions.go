package api

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/api/route/contact"

	"github.com/cygy/ginamite/api/cache"
	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/api/route/terms"
	"github.com/cygy/ginamite/api/route/user"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"
	mongo "github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
	"github.com/cygy/ginamite/common/weburl"

	"github.com/fvbock/endless"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CreateKafkaConsumers : returns an array of kafka consumers.
func (s *Server) CreateKafkaConsumers() []kafka.MessageConsumer {
	consumers := []kafka.MessageConsumer{}

	// Handle the tasks.
	consumers = append(consumers, kafka.MessageConsumer{
		Topic: queue.TopicCache,
		Action: func(value []byte) {
			defer kafka.Recover()

			msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicCache)
			if !ok {
				return
			}

			switch msgType {
			case queue.MessageCacheInvalidAuthToken:
				p := queue.CacheAuthToken{}
				queue.UnmarshalPayload(payload, &p)
				cache.AddRevokedAuthTokens(p.Tokens)
				break
			default:
				if s.Functions.NewMessageFromQueueCache == nil || !s.Functions.NewMessageFromQueueCache(msgType) {
					log.WithFields(logrus.Fields{
						"topic": queue.TopicCache,
						"type":  msgType,
					}).Warn("unknown kafka message")
				}
			}

		}})

	return consumers
}

// Start : starts the worker.
func (s *Server) Start() {
	// Warn about the undefined functions.
	s.warnUndefinedFunctions()

	// Enable debug mode if needed.
	if config.Main.IsDebugModeEnabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create the server.
	var server *gin.Engine

	if config.Main.IsProductionEnvironment() {
		server = gin.New()
		server.Use(gin.Recovery())
		server.Use(log.InjectRequestLogger(true))
	} else {
		server = gin.Default()
	}

	// Enable server settings
	if config.Main.Server.GZip.Enabled {
		server.Use(gzip.Gzip(config.Main.Server.GZip.Level))
	}

	// Set up some global functions.
	authentication.BuildTokenFromID = s.Functions.BuildTokenFromID
	authentication.ExtraPropertiesForTokenWithID = s.Functions.ExtraPropertiesForTokenWithID
	authentication.ExtendTokenExpirationDateFromID = s.Functions.ExtendTokenExpirationDateFromID
	terms.UpdateVersionOfTermsAcceptedByUser = s.Functions.UpdateVersionOfTermsAcceptedByUser
	user.GetEnabledUsersForAdmin = s.Functions.GetEnabledUsersForAdmin
	user.GetDisabledUsersForAdmin = s.Functions.GetDisabledUsersForAdmin
	user.UpdatePasswordByAdmin = s.Functions.UpdatePasswordByAdmin
	user.UpdateInformationsByAdmin = s.Functions.UpdateInformationsByAdmin
	user.UpdateDescriptionByAdmin = s.Functions.UpdateDescriptionByAdmin
	user.UpdateSocialByAdmin = s.Functions.UpdateSocialByAdmin
	user.UpdateSettingsByAdmin = s.Functions.UpdateSettingsByAdmin
	user.UpdateNotificationsByAdmin = s.Functions.UpdateNotificationsByAdmin
	user.GetTokensForAdmin = s.Functions.GetTokensForAdmin
	user.DeleteTokenByIDByAdmin = s.Functions.DeleteTokenByIDByAdmin
	user.SetAbilitiesToUser = s.Functions.SetAbilitiesToUser
	user.AddAbilitiesToUser = s.Functions.AddAbilitiesToUser
	user.RemoveAbilitiesFromUser = s.Functions.RemoveAbilitiesFromUser
	account.NotificationsReceivedByUser = s.Functions.NotificationsReceivedByUser
	account.NotificationReceivedByUser = s.Functions.NotificationReceivedByUser
	account.UpdateUserNotificationRead = s.Functions.UpdateUserNotificationRead
	account.UpdateUserNotificationNotified = s.Functions.UpdateUserNotificationNotified
	account.DeleteUserNotification = s.Functions.DeleteUserNotification
	account.NotificationsSubscribedByUser = s.Functions.NotificationsSubscribedByUser
	account.SubscribeNotificationsToUser = s.Functions.SubscribeNotificationsToUser
	account.UnsubscribeNotificationFromUser = s.Functions.UnsubscribeNotificationFromUser
	account.GetUserDetailsByIdentifier = s.Functions.GetUserDetailsByIdentifier
	account.GetUserDetailsByID = s.Functions.GetUserDetailsByID
	account.NewForgotPasswordProcessForUser = s.Functions.NewForgotPasswordProcessForUser
	account.GetForgotPasswordProcessDetailsByID = s.Functions.GetForgotPasswordProcessDetailsByID
	account.WrongAccessToForgotPasswordProcess = s.Functions.WrongAccessToForgotPasswordProcess
	account.UpdateUserPassword = s.Functions.UpdateUserPassword
	account.UpdateUserEmailAddress = s.Functions.UpdateUserEmailAddress
	account.UserEmailAddressIsVerified = s.Functions.UserEmailAddressIsVerified
	account.UpdateUserSocialNetworks = s.Functions.UpdateUserSocialNetworks
	account.DeleteForgotPasswordProcess = s.Functions.DeleteForgotPasswordProcess
	account.ValidateUserRegistration = s.Functions.ValidateUserRegistration
	account.CancelUserRegistration = s.Functions.CancelUserRegistration
	account.GetRegistrationDetails = s.Functions.GetRegistrationDetails
	account.SaveAuthenticationToken = s.Functions.SaveAuthenticationToken
	account.DoesAccountWithUsernameExist = s.Functions.DoesAccountWithUsernameExist
	account.DoesAccountWithEmailAddressExist = s.Functions.DoesAccountWithEmailAddressExist
	account.RegisterByEmailAddress = s.Functions.RegisterByEmailAddress
	account.RegisterByThirdPartyToken = s.Functions.RegisterByThirdPartyToken
	account.DoesUserWithFacebookUserIDExist = s.Functions.DoesUserWithFacebookUserIDExist
	account.DoesUserWithGoogleUserIDExist = s.Functions.DoesUserWithGoogleUserIDExist
	account.GetUserDetailsByFacebookUserID = s.Functions.GetUserDetailsByFacebookUserID
	account.GetUserDetailsByGoogleUserID = s.Functions.GetUserDetailsByGoogleUserID
	account.SaveFacebookUserTokenDetailsToUser = s.Functions.SaveFacebookUserTokenDetailsToUser
	account.SaveGoogleUserTokenDetailsToUser = s.Functions.SaveGoogleUserTokenDetailsToUser
	account.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser = s.Functions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser
	account.GetTokenExpirationDate = s.Functions.GetTokenExpirationDate
	account.DeleteTokenByID = s.Functions.DeleteTokenByID
	account.GetTokenDetailsByID = s.Functions.GetTokenDetailsByID
	account.GetOwnedTokens = s.Functions.GetOwnedTokens
	account.GetOwnedTokenByID = s.Functions.GetOwnedTokenByID
	account.UpdateTokenByID = s.Functions.UpdateTokenByID
	account.GetOwnedAccountDetails = s.Functions.GetOwnedAccountDetails
	account.UpdatedAccountInfos = s.Functions.UpdatedAccountInfos
	account.UpdatedAccountSettings = s.Functions.UpdatedAccountSettings
	account.CreateNewDeleteAccountProcess = s.Functions.CreateNewDeleteAccountProcess
	account.CreateNewDisableAccountProcess = s.Functions.CreateNewDisableAccountProcess
	account.CreateNewUpdatePasswordProcess = s.Functions.CreateNewUpdatePasswordProcess
	account.CreateNewUpdateEmailAddressProcess = s.Functions.CreateNewUpdateEmailAddressProcess
	account.CreateNewVerifyEmailAddressProcess = s.Functions.CreateNewVerifyEmailAddressProcess
	account.VerifyProcess = s.Functions.VerifyProcess
	account.DeleteProcessByID = s.Functions.DeleteProcessByID
	account.DeleteUserByID = s.Functions.DeleteUserByID
	account.DisableUserByID = s.Functions.DisableUserByID
	account.EnableUserByID = s.Functions.EnableUserByID
	contact.CreateContact = s.Functions.CreateContact
	contact.UpdateContact = s.Functions.UpdateContact
	contact.GetContacts = s.Functions.GetContacts
	contact.GetContact = s.Functions.GetContact

	// Load the middlewares.
	if config.Main.IsMongoDBDatabase() {
		server.Use(mongo.Inject())
	}
	server.Use(context.InjectLocale(config.Main.DefaultLocale, config.Main.SupportedLocales))

	// Load the templates.
	html.LoadTemplatesWithDelimiters(config.Main.TemplatesPath+"/html", "[[", "]]", server)
	text.LoadTemplates(config.Main.TemplatesPath + "/text")

	// Initialize the configuration of the HTML routes.
	weburl.Initialize(s.RoutesFile, config.Main.Hosts.Web)

	// Set up the routes.
	router := route.NewRouter(
		server, config.Main.APIVersion, config.Main.IsProductionEnvironment(), config.Main.JWT.Secret, config.Main.Terms.Version,
		s.Functions.MiddlewareIsValidAuthToken,
		s.Functions.MiddlewareGetUserAndLocaleFromAuthToken,
		s.Functions.MiddlewareGetUserAbilities,
		s.Functions.MiddlewareGetLatestVersionOfTermsAcceptedByUser,
	)
	if s.Functions.RouterConfigureDefaultRoutes != nil {
		s.Functions.RouterConfigureDefaultRoutes(&router.DefaultRoutes)
	}
	router.LoadDefaultRoutes()
	for _, customRoute := range s.customRoutes {
		customRoute(router.Group, router.Handlers, router.ProductionEnvironment, router.Rights)
	}

	// Start the built-in recurring tasks.
	cache.StartDeletingExpiredTokens(time.Duration(config.Main.ExpiredAuthTokens.FlushCacheInterval))

	// Start the custom recurring tasks.
	if s.Functions.RecurringTasks != nil {
		s.Functions.RecurringTasks()
	}

	// Start the server.
	address := fmt.Sprintf(":%d", config.Main.Server.Port)
	if config.Main.Server.SSL.Enabled {
		endless.ListenAndServeTLS(address, config.Main.Server.SSL.CertPath, config.Main.Server.SSL.KeyPath, server)
	} else {
		endless.ListenAndServe(address, server)
	}
}

// AddCustomRoutes : adds some suctom routes.
func (s *Server) AddCustomRoutes(configuration route.ConfigureCustomRoutes) {
	s.customRoutes = append(s.customRoutes, configuration)
}

// NewServer : returns a new struct 'Server'.
func NewServer() *Server {
	return &Server{
		customRoutes: []route.ConfigureCustomRoutes{},
	}
}

// Helping functions.
func (s *Server) warnUndefinedFunctions() {
	if s.Functions.RecurringTasks == nil {
		log.Warn("The function 'RecurringTasks' is undefined.")
	}
	if s.Functions.NewMessageFromQueueCache == nil {
		log.Warn("The function 'NewMessageFromQueueCache' is undefined.")
	}
	if s.Functions.MiddlewareIsValidAuthToken == nil {
		log.Warn("The function 'MiddlewareIsValidAuthToken' is undefined.")
	}
	if s.Functions.MiddlewareGetUserAndLocaleFromAuthToken == nil {
		log.Warn("The function 'MiddlewareGetUserAndLocaleFromAuthToken' is undefined.")
	}
	if s.Functions.MiddlewareGetUserAbilities == nil {
		log.Warn("The function 'MiddlewareGetUserAbilities' is undefined.")
	}
	if s.Functions.MiddlewareGetLatestVersionOfTermsAcceptedByUser == nil {
		log.Warn("The function 'MiddlewareGetLatestVersionOfTermsAcceptedByUser' is undefined.")
	}
	if s.Functions.BuildTokenFromID == nil {
		log.Warn("The function 'BuildTokenFromID' is undefined.")
	}
	if s.Functions.ExtraPropertiesForTokenWithID == nil {
		log.Warn("The function 'ExtraPropertiesForTokenWithID' is undefined.")
	}
	if s.Functions.ExtendTokenExpirationDateFromID == nil {
		log.Warn("The function 'ExtendTokenExpirationDateFromID' is undefined.")
	}
	if s.Functions.RouterConfigureDefaultRoutes == nil {
		log.Warn("The function 'RouterConfigureDefaultRoutes' is undefined.")
	}
	if s.Functions.UpdateVersionOfTermsAcceptedByUser == nil {
		log.Warn("The function 'UpdateVersionOfTermsAcceptedByUser' is undefined.")
	}
	if s.Functions.GetEnabledUsersForAdmin == nil {
		log.Warn("The function 'GetEnabledUsersForAdmin' is undefined.")
	}
	if s.Functions.GetDisabledUsersForAdmin == nil {
		log.Warn("The function 'GetDisabledUsersForAdmin' is undefined.")
	}
	if s.Functions.UpdatePasswordByAdmin == nil {
		log.Warn("The function 'UpdatePasswordByAdmin' is undefined.")
	}
	if s.Functions.UpdateInformationsByAdmin == nil {
		log.Warn("The function 'UpdateInformationsByAdmin' is undefined.")
	}
	if s.Functions.UpdateDescriptionByAdmin == nil {
		log.Warn("The function 'UpdateDescriptionByAdmin' is undefined.")
	}
	if s.Functions.UpdateSocialByAdmin == nil {
		log.Warn("The function 'UpdateSocialByAdmin' is undefined.")
	}
	if s.Functions.UpdateSettingsByAdmin == nil {
		log.Warn("The function 'UpdateSettingsByAdmin' is undefined.")
	}
	if s.Functions.UpdateNotificationsByAdmin == nil {
		log.Warn("The function 'UpdateNotificationsByAdmin' is undefined.")
	}
	if s.Functions.GetTokensForAdmin == nil {
		log.Warn("The function 'GetTokensForAdmin' is undefined.")
	}
	if s.Functions.DeleteTokenByIDByAdmin == nil {
		log.Warn("The function 'DeleteTokenByIDByAdmin' is undefined.")
	}
	if s.Functions.SetAbilitiesToUser == nil {
		log.Warn("The function 'SetAbilitiesToUser' is undefined.")
	}
	if s.Functions.AddAbilitiesToUser == nil {
		log.Warn("The function 'AddAbilitiesToUser' is undefined.")
	}
	if s.Functions.RemoveAbilitiesFromUser == nil {
		log.Warn("The function 'RemoveAbilitiesFromUser' is undefined.")
	}
	if s.Functions.NotificationsReceivedByUser == nil {
		log.Warn("The function 'NotificationsReceivedByUser' is undefined.")
	}
	if s.Functions.NotificationReceivedByUser == nil {
		log.Warn("The function 'NotificationReceivedByUser' is undefined.")
	}
	if s.Functions.UpdateUserNotificationRead == nil {
		log.Warn("The function 'UpdateUserNotificationRead' is undefined.")
	}
	if s.Functions.UpdateUserNotificationNotified == nil {
		log.Warn("The function 'UpdateUserNotificationNotified' is undefined.")
	}
	if s.Functions.DeleteUserNotification == nil {
		log.Warn("The function 'DeleteUserNotification' is undefined.")
	}
	if s.Functions.NotificationsSubscribedByUser == nil {
		log.Warn("The function 'NotificationsSubscribedByUser' is undefined.")
	}
	if s.Functions.SubscribeNotificationsToUser == nil {
		log.Warn("The function 'SubscribeNotificationsToUser' is undefined.")
	}
	if s.Functions.UnsubscribeNotificationFromUser == nil {
		log.Warn("The function 'UnsubscribeNotificationFromUser' is undefined.")
	}
	if s.Functions.GetUserDetailsByIdentifier == nil {
		log.Warn("The function 'GetUserDetailsByIdentifier' is undefined.")
	}
	if s.Functions.GetUserDetailsByID == nil {
		log.Warn("The function 'GetUserDetailsByID' is undefined.")
	}
	if s.Functions.NewForgotPasswordProcessForUser == nil {
		log.Warn("The function 'NewForgotPasswordProcessForUser' is undefined.")
	}
	if s.Functions.GetForgotPasswordProcessDetailsByID == nil {
		log.Warn("The function 'GetForgotPasswordProcessDetailsByID' is undefined.")
	}
	if s.Functions.WrongAccessToForgotPasswordProcess == nil {
		log.Warn("The function 'WrongAccessToForgotPasswordProcess' is undefined.")
	}
	if s.Functions.UpdateUserPassword == nil {
		log.Warn("The function 'UpdateUserPassword' is undefined.")
	}
	if s.Functions.UpdateUserEmailAddress == nil {
		log.Warn("The function 'UpdateUserEmailAddress' is undefined.")
	}
	if s.Functions.UserEmailAddressIsVerified == nil {
		log.Warn("The function 'UserEmailAddressIsVerified' is undefined.")
	}
	if s.Functions.UpdateUserSocialNetworks == nil {
		log.Warn("The function 'UpdateUserSocialNetworks' is undefined.")
	}
	if s.Functions.DeleteForgotPasswordProcess == nil {
		log.Warn("The function 'DeleteForgotPasswordProcess' is undefined.")
	}
	if s.Functions.ValidateUserRegistration == nil {
		log.Warn("The function 'ValidateUserRegistration' is undefined.")
	}
	if s.Functions.CancelUserRegistration == nil {
		log.Warn("The function 'CancelUserRegistration' is undefined.")
	}
	if s.Functions.GetRegistrationDetails == nil {
		log.Warn("The function 'GetRegistrationDetails' is undefined.")
	}
	if s.Functions.SaveAuthenticationToken == nil {
		log.Warn("The function 'SaveAuthenticationToken' is undefined.")
	}
	if s.Functions.DoesAccountWithUsernameExist == nil {
		log.Warn("The function 'DoesAccountWithUsernameExist' is undefined.")
	}
	if s.Functions.DoesAccountWithEmailAddressExist == nil {
		log.Warn("The function 'DoesAccountWithEmailAddressExist' is undefined.")
	}
	if s.Functions.RegisterByEmailAddress == nil {
		log.Warn("The function 'RegisterByEmailAddress' is undefined.")
	}
	if s.Functions.RegisterByThirdPartyToken == nil {
		log.Warn("The function 'RegisterByThirdPartyToken' is undefined.")
	}
	if s.Functions.DoesUserWithFacebookUserIDExist == nil {
		log.Warn("The function 'DoesUserWithFacebookUserIDExist' is undefined.")
	}
	if s.Functions.DoesUserWithGoogleUserIDExist == nil {
		log.Warn("The function 'DoesUserWithGoogleUserIDExist' is undefined.")
	}
	if s.Functions.GetUserDetailsByFacebookUserID == nil {
		log.Warn("The function 'GetUserDetailsByFacebookUserID' is undefined.")
	}
	if s.Functions.GetUserDetailsByGoogleUserID == nil {
		log.Warn("The function 'GetUserDetailsByGoogleUserID' is undefined.")
	}
	if s.Functions.SaveFacebookUserTokenDetailsToUser == nil {
		log.Warn("The function 'SaveFacebookUserTokenDetailsToUser' is undefined.")
	}
	if s.Functions.SaveGoogleUserTokenDetailsToUser == nil {
		log.Warn("The function 'SaveGoogleUserTokenDetailsToUser' is undefined.")
	}
	if s.Functions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser == nil {
		log.Warn("The function 'GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser' is undefined.")
	}
	if s.Functions.GetTokenExpirationDate == nil {
		log.Warn("The function 'GetTokenExpirationDate' is undefined.")
	}
	if s.Functions.DeleteTokenByID == nil {
		log.Warn("The function 'DeleteTokenByID' is undefined.")
	}
	if s.Functions.GetTokenDetailsByID == nil {
		log.Warn("The function 'GetTokenDetailsByID' is undefined.")
	}
	if s.Functions.GetOwnedTokens == nil {
		log.Warn("The function 'GetOwnedTokens' is undefined.")
	}
	if s.Functions.GetOwnedTokenByID == nil {
		log.Warn("The function 'GetOwnedTokenByID' is undefined.")
	}
	if s.Functions.UpdateTokenByID == nil {
		log.Warn("The function 'UpdateTokenByID' is undefined.")
	}
	if s.Functions.GetOwnedAccountDetails == nil {
		log.Warn("The function 'GetOwnedAccountDetails' is undefined.")
	}
	if s.Functions.UpdatedAccountInfos == nil {
		log.Warn("The function 'UpdatedAccountInfos' is undefined.")
	}
	if s.Functions.UpdatedAccountSettings == nil {
		log.Warn("The function 'UpdatedAccountSettings' is undefined.")
	}
	if s.Functions.CreateNewDeleteAccountProcess == nil {
		log.Warn("The function 'CreateNewDeleteAccountProcess' is undefined.")
	}
	if s.Functions.CreateNewDisableAccountProcess == nil {
		log.Warn("The function 'CreateNewDisableAccountProcess' is undefined.")
	}
	if s.Functions.CreateNewUpdatePasswordProcess == nil {
		log.Warn("The function 'CreateNewUpdatePasswordProcess' is undefined.")
	}
	if s.Functions.CreateNewUpdateEmailAddressProcess == nil {
		log.Warn("The function 'CreateNewUpdateEmailAddressProcess' is undefined.")
	}
	if s.Functions.CreateNewVerifyEmailAddressProcess == nil {
		log.Warn("The function 'CreateNewVerifyEmailAddressProcess' is undefined.")
	}
	if s.Functions.VerifyProcess == nil {
		log.Warn("The function 'VerifyProcess' is undefined.")
	}
	if s.Functions.DeleteProcessByID == nil {
		log.Warn("The function 'DeleteProcessByID' is undefined.")
	}
	if s.Functions.DeleteUserByID == nil {
		log.Warn("The function 'DeleteUserByID' is undefined.")
	}
	if s.Functions.DisableUserByID == nil {
		log.Warn("The function 'DisableUserByID' is undefined.")
	}
	if s.Functions.EnableUserByID == nil {
		log.Warn("The function 'EnableUserByID' is undefined.")
	}
	if s.Functions.CreateContact == nil {
		log.Warn("The function 'CreateContact' is undefined.")
	}
	if s.Functions.UpdateContact == nil {
		log.Warn("The function 'UpdateContact' is undefined.")
	}
	if s.Functions.GetContacts == nil {
		log.Warn("The function 'GetContacts' is undefined.")
	}
	if s.Functions.GetContact == nil {
		log.Warn("The function 'GetContact' is undefined.")
	}
}
