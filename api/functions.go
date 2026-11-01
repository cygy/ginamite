// Package api wires the HTTP API server: routes, middlewares and the functions provided by the application.
package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/cygy/ginamite/api/route/contact"

	tokencache "github.com/cygy/ginamite/api/cache/authtoken"
	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/api/route/terms"
	"github.com/cygy/ginamite/api/route/user"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/ctxlocalization"
	"github.com/cygy/ginamite/common/ctxnotifications"
	"github.com/cygy/ginamite/common/ctxqueue"
	"github.com/cygy/ginamite/common/ctxtimezone"
	"github.com/cygy/ginamite/common/httpserver"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// errUndecodableCache is the error of a message of the cache which can not be decoded, which no retry can fix.
var errUndecodableCache = broker.Permanent(errors.New("undecodable message"))

// CreateConsumers returns an array of kafka consumers.
func (s *Server) CreateConsumers() []broker.MessageConsumer {
	consumers := []broker.MessageConsumer{}

	// Handle the messages of the cache. Every instance of the API must receive every message, so each one has its own
	// group. A message which can not be decoded is dead-lettered.
	consumers = append(consumers, broker.MessageConsumer{
		Topic:      queue.TopicCache,
		Group:      broker.UniqueGroup(queue.TopicCache),
		Idempotent: true,
		Action: func(value []byte) error {
			msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicCache)
			if !ok {
				return errUndecodableCache
			}

			switch msgType {
			case queue.MessageCacheInvalidAuthToken:
				p := queue.CacheAuthToken{}
				if err := queue.UnmarshalPayload(payload, &p); err != nil {
					return broker.Permanent(err)
				}
				s.RevokedTokens.Add(p.Tokens)
			default:
				if s.Functions.NewMessageFromQueueCache == nil || !s.Functions.NewMessageFromQueueCache(msgType) {
					log.WithFields(logrus.Fields{
						"topic": queue.TopicCache,
						"type":  msgType,
					}).Warn("unknown kafka message")
				}
			}

			return nil
		},
	})

	return consumers
}

// Start starts the worker.
func (s *Server) Start() {
	// Warn about the undefined functions.
	s.warnUndefinedFunctions()
	s.warnWeakSecret()

	// Enable debug mode if needed.
	if s.Config.IsDebugModeEnabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create the server.
	var server *gin.Engine

	if s.Config.IsProductionEnvironment() {
		server = gin.New()
		server.Use(gin.Recovery())
		server.Use(log.InjectRequestLogger(true))
	} else {
		server = gin.Default()
	}

	s.configureProxies(server)

	// Enable server settings
	if s.Config.Server.GZip.Enabled {
		server.Use(gzip.Gzip(s.Config.Server.GZip.Level))
	}

	// Set up some global functions.
	accountFunctions, userFunctions, contactFunctions, termsFunctions := s.routeFunctions()

	// Load the middlewares.
	if s.Requests != nil {
		server.Use(s.Requests.Middleware())
	}
	server.Use(ctxconfig.Inject(s.Config), ctxqueue.Inject(s.Queue), ctxlocalization.Inject(s.Localization), ctxtimezone.Inject(s.Timezones), ctxnotifications.Inject(s.Notifications))
	server.Use(s.Middlewares...)
	server.Use(s.idValidator())
	server.Use(ctxAuthToken.InjectRevoked(s.RevokedTokens.IsRevoked))
	server.Use(ctxAuthToken.InjectHooks(authentication.Hooks{
		BuildTokenFromID:                s.Functions.BuildTokenFromID,
		ExtraPropertiesForTokenWithID:   s.Functions.ExtraPropertiesForTokenWithID,
		ExtendTokenExpirationDateFromID: s.Functions.ExtendTokenExpirationDateFromID,
	}))
	server.Use(
		account.Inject(&accountFunctions),
		user.Inject(&userFunctions),
		contact.Inject(&contactFunctions),
		terms.Inject(&termsFunctions),
	)
	server.Use(ctxLocale.Inject(s.Config.DefaultLocale, s.Config.SupportedLocales))

	// Load the HTML templates into the engine, for the custom routes of the application.
	if s.Config.TemplatesPath != "" {
		html.LoadTemplatesWithDelimiters(s.Config.TemplatesPath+"/html", "[[", "]]", server)
	}

	// Set up the routes.
	router := route.NewRouter(
		server, s.Config.APIVersion, s.Config.IsProductionEnvironment(), s.Config.JWT.Secret, s.Config.Terms.Version,
		s.Functions.MiddlewareIsValidAuthToken,
		s.Functions.MiddlewareGetUserAndLocaleFromAuthToken,
		s.Functions.MiddlewareGetUserAbilities,
		s.Functions.MiddlewareGetLatestVersionOfTermsAcceptedByUser,
	)
	if s.Functions.RouterConfigureDefaultRoutes != nil {
		s.Functions.RouterConfigureDefaultRoutes(&router.DefaultRoutes)
	}
	router.Notifications = s.Notifications
	router.Limits = route.RateLimits{
		Authentication: request.RateLimitFromConfiguration(s.Config.RateLimits.Authentication),
		Registration:   request.RateLimitFromConfiguration(s.Config.RateLimits.Registration),
		ForgotPassword: request.RateLimitFromConfiguration(s.Config.RateLimits.ForgotPassword),
	}
	router.LoadDefaultRoutes()
	for _, customRoute := range s.customRoutes {
		customRoute(router.Group, router.Handlers, router.ProductionEnvironment, router.Rights)
	}

	// Start the built-in recurring tasks.
	s.RevokedTokens.StartDeletingExpired(time.Duration(s.Config.ExpiredAuthTokens.FlushCacheInterval))

	// Start the custom recurring tasks.
	if s.Functions.RecurringTasks != nil {
		s.Functions.RecurringTasks()
	}

	// Start the server.
	address := fmt.Sprintf(":%d", s.Config.Server.Port)
	if s.Config.Server.SSL.Enabled {
		if err := httpserver.ListenAndServeTLS(address, s.Config.Server.SSL.CertPath, s.Config.Server.SSL.KeyPath, server); err != nil {
			log.WithField("error", err).Error("the server stopped")
		}
	} else if err := httpserver.ListenAndServe(address, server); err != nil {
		log.WithField("error", err).Error("the server stopped")
	}
}

// AddCustomRoutes adds some suctom routes.
func (s *Server) AddCustomRoutes(configuration route.ConfigureCustomRoutes) {
	s.customRoutes = append(s.customRoutes, configuration)
}

// NewServer returns a new struct 'Server'.
func NewServer() *Server {
	return &Server{
		Config:        config.NewConfiguration(),
		Localization:  localization.NewEngine(),
		Timezones:     timezone.NewConfiguration(),
		Notifications: notifications.NewRegistry(),
		Queue:         queue.New(nil),
		RevokedTokens: tokencache.NewRevoked(),
		customRoutes:  []route.ConfigureCustomRoutes{},
	}
}

// routeFunctions returns the functions of the server used by the handlers of the routes, by package.
func (s *Server) routeFunctions() (accountFunctions account.Functions, userFunctions user.Functions, contactFunctions contact.Functions, termsFunctions terms.Functions) {
	termsFunctions.UpdateVersionOfTermsAcceptedByUser = s.Functions.UpdateVersionOfTermsAcceptedByUser
	userFunctions.GetEnabledUsersForAdmin = s.Functions.GetEnabledUsersForAdmin
	userFunctions.GetDisabledUsersForAdmin = s.Functions.GetDisabledUsersForAdmin
	userFunctions.UpdatePasswordByAdmin = s.Functions.UpdatePasswordByAdmin
	userFunctions.UpdateInformationsByAdmin = s.Functions.UpdateInformationsByAdmin
	userFunctions.UpdateDescriptionByAdmin = s.Functions.UpdateDescriptionByAdmin
	userFunctions.UpdateSocialByAdmin = s.Functions.UpdateSocialByAdmin
	userFunctions.UpdateSettingsByAdmin = s.Functions.UpdateSettingsByAdmin
	userFunctions.UpdateNotificationsByAdmin = s.Functions.UpdateNotificationsByAdmin
	userFunctions.GetTokensForAdmin = s.Functions.GetTokensForAdmin
	userFunctions.DeleteTokenByIDByAdmin = s.Functions.DeleteTokenByIDByAdmin
	userFunctions.SetAbilitiesToUser = s.Functions.SetAbilitiesToUser
	userFunctions.AddAbilitiesToUser = s.Functions.AddAbilitiesToUser
	userFunctions.RemoveAbilitiesFromUser = s.Functions.RemoveAbilitiesFromUser
	accountFunctions.NotificationsReceivedByUser = s.Functions.NotificationsReceivedByUser
	accountFunctions.NotificationReceivedByUser = s.Functions.NotificationReceivedByUser
	accountFunctions.UpdateUserNotificationRead = s.Functions.UpdateUserNotificationRead
	accountFunctions.UpdateUserNotificationNotified = s.Functions.UpdateUserNotificationNotified
	accountFunctions.DeleteUserNotification = s.Functions.DeleteUserNotification
	accountFunctions.NotificationsSubscribedByUser = s.Functions.NotificationsSubscribedByUser
	accountFunctions.SubscribeNotificationsToUser = s.Functions.SubscribeNotificationsToUser
	accountFunctions.UnsubscribeNotificationFromUser = s.Functions.UnsubscribeNotificationFromUser
	accountFunctions.GetUserDetailsByIdentifier = s.Functions.GetUserDetailsByIdentifier
	accountFunctions.GetUserDetailsByID = s.Functions.GetUserDetailsByID
	accountFunctions.NewForgotPasswordProcessForUser = s.Functions.NewForgotPasswordProcessForUser
	accountFunctions.GetForgotPasswordProcessDetailsByID = s.Functions.GetForgotPasswordProcessDetailsByID
	accountFunctions.WrongAccessToForgotPasswordProcess = s.Functions.WrongAccessToForgotPasswordProcess
	accountFunctions.UpdateUserPassword = s.Functions.UpdateUserPassword
	accountFunctions.UpdateUserEmailAddress = s.Functions.UpdateUserEmailAddress
	accountFunctions.UserEmailAddressIsVerified = s.Functions.UserEmailAddressIsVerified
	accountFunctions.UpdateUserSocialNetworks = s.Functions.UpdateUserSocialNetworks
	accountFunctions.DeleteForgotPasswordProcess = s.Functions.DeleteForgotPasswordProcess
	accountFunctions.ValidateUserRegistration = s.Functions.ValidateUserRegistration
	accountFunctions.CancelUserRegistration = s.Functions.CancelUserRegistration
	accountFunctions.GetRegistrationDetails = s.Functions.GetRegistrationDetails
	accountFunctions.SaveAuthenticationToken = s.Functions.SaveAuthenticationToken
	accountFunctions.DoesAccountWithUsernameExist = s.Functions.DoesAccountWithUsernameExist
	accountFunctions.DoesAccountWithEmailAddressExist = s.Functions.DoesAccountWithEmailAddressExist
	accountFunctions.RegisterByEmailAddress = s.Functions.RegisterByEmailAddress
	accountFunctions.RegisterByThirdPartyToken = s.Functions.RegisterByThirdPartyToken
	accountFunctions.DoesUserWithFacebookUserIDExist = s.Functions.DoesUserWithFacebookUserIDExist
	accountFunctions.DoesUserWithGoogleUserIDExist = s.Functions.DoesUserWithGoogleUserIDExist
	accountFunctions.GetUserDetailsByFacebookUserID = s.Functions.GetUserDetailsByFacebookUserID
	accountFunctions.GetUserDetailsByGoogleUserID = s.Functions.GetUserDetailsByGoogleUserID
	accountFunctions.SaveFacebookUserTokenDetailsToUser = s.Functions.SaveFacebookUserTokenDetailsToUser
	accountFunctions.SaveGoogleUserTokenDetailsToUser = s.Functions.SaveGoogleUserTokenDetailsToUser
	accountFunctions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser = s.Functions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser
	accountFunctions.UpdateEncryptedPasswordByUserID = s.Functions.UpdateEncryptedPasswordByUserID
	accountFunctions.GetTokenExpirationDate = s.Functions.GetTokenExpirationDate
	accountFunctions.DeleteTokenByID = s.Functions.DeleteTokenByID
	accountFunctions.GetTokenDetailsByID = s.Functions.GetTokenDetailsByID
	accountFunctions.GetOwnedTokens = s.Functions.GetOwnedTokens
	accountFunctions.GetOwnedTokenByID = s.Functions.GetOwnedTokenByID
	accountFunctions.UpdateTokenByID = s.Functions.UpdateTokenByID
	accountFunctions.GetOwnedAccountDetails = s.Functions.GetOwnedAccountDetails
	accountFunctions.UpdatedAccountInfos = s.Functions.UpdatedAccountInfos
	accountFunctions.UpdatedAccountSettings = s.Functions.UpdatedAccountSettings
	accountFunctions.CreateNewDeleteAccountProcess = s.Functions.CreateNewDeleteAccountProcess
	accountFunctions.CreateNewDisableAccountProcess = s.Functions.CreateNewDisableAccountProcess
	accountFunctions.CreateNewUpdatePasswordProcess = s.Functions.CreateNewUpdatePasswordProcess
	accountFunctions.CreateNewUpdateEmailAddressProcess = s.Functions.CreateNewUpdateEmailAddressProcess
	accountFunctions.CreateNewVerifyEmailAddressProcess = s.Functions.CreateNewVerifyEmailAddressProcess
	accountFunctions.VerifyProcess = s.Functions.VerifyProcess
	accountFunctions.DeleteProcessByID = s.Functions.DeleteProcessByID
	accountFunctions.DeleteUserByID = s.Functions.DeleteUserByID
	accountFunctions.DisableUserByID = s.Functions.DisableUserByID
	accountFunctions.EnableUserByID = s.Functions.EnableUserByID
	contactFunctions.CreateContact = s.Functions.CreateContact
	contactFunctions.UpdateContact = s.Functions.UpdateContact
	contactFunctions.GetContacts = s.Functions.GetContacts
	contactFunctions.GetContact = s.Functions.GetContact

	return accountFunctions, userFunctions, contactFunctions, termsFunctions
}

// idValidator returns the middleware making the requests use the IsValidID function of the server, if any, to validate
// the identifiers. It does nothing if the function is undefined.
func (s *Server) idValidator() gin.HandlerFunc {
	if s.Functions.IsValidID == nil {
		return func(c *gin.Context) {}
	}

	return request.InjectIDValidator(s.Functions.IsValidID)
}

// Helping functions.
func (s *Server) warnUndefinedFunctions() {
	if s.Functions.RecurringTasks == nil {
		log.Warn("The function 'RecurringTasks' is undefined.")
	}
	if s.Functions.IsValidID == nil {
		log.Warn("The function 'IsValidID' is undefined: every identifier will be accepted.")
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
	if s.Functions.UpdateEncryptedPasswordByUserID == nil {
		log.Warn("The function 'UpdateEncryptedPasswordByUserID' is undefined: the weak password hashes are not upgraded.")
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

// configureProxies sets the proxies trusted to give the address of the client, which limits the requests. Gin trusts
// every proxy by default, so a client can forge the address: a warning asks to configure them in production.
func (s *Server) configureProxies(server *gin.Engine) {
	if s.Config.Server.TrustedProxies == nil {
		if s.Config.IsProductionEnvironment() {
			log.Warn("The setting 'server.trusted_proxies' is undefined: every proxy is trusted, the address of the clients can be forged.")
		}

		return
	}

	if err := server.SetTrustedProxies(s.Config.Server.TrustedProxies); err != nil {
		log.WithField("error", err).Fatal("invalid setting 'server.trusted_proxies'")
	}
}

// minimumSecretLength is the length under which the secret of the JWT can be guessed.
const minimumSecretLength = 32

// warnWeakSecret logs a warning if the secret signing the auth tokens is short.
func (s *Server) warnWeakSecret() {
	if len(s.Config.JWT.Secret) < minimumSecretLength {
		log.Warn(fmt.Sprintf("The setting 'jwt.secret' is shorter than %d characters: the auth tokens can be forged.", minimumSecretLength))
	}
}
