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
	"github.com/cygy/ginamite/common/ctxkey"
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

// Run sets up the routes, then serves the API until SIGINT or SIGTERM. It returns an error if the set up fails or if
// the server stops for another reason.
func (s *Server) Run() error {
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

	if err := s.configureProxies(server); err != nil {
		return err
	}

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
	server.Use(ctxkey.Config.Inject(s.Config), ctxkey.Queue.Inject(s.Queue), ctxkey.Localization.Inject(s.Localization), ctxkey.Timezones.Inject(s.Timezones), ctxkey.Notifications.Inject(s.Notifications))
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
		if _, err := html.LoadTemplatesWithDelimiters(s.Config.TemplatesPath+"/html", "[[", "]]", server); err != nil {
			return err
		}
	}

	// Set up the routes.
	router := route.NewRouter(
		server, s.Config.APIVersion, s.Config.IsProductionEnvironment(), s.Config.JWT.Secret, s.Config.Terms.Version,
		s.Functions.MiddlewareIsValidAuthToken,
		s.Functions.MiddlewareGetUserAndLocaleFromAuthToken,
		s.Functions.MiddlewareGetUserAbilities,
		s.Functions.MiddlewareGetLatestVersionOfTermsAcceptedByUser,
	)
	needsJWT, err := s.selectRoutes(&router.DefaultRoutes, s.Config.APIFeatures)
	if err != nil {
		return fmt.Errorf("unable to load the routes of the API: %w", err)
	}
	if router.DefaultRoutes.Authentication.Paths.LoginWithPassword.Enabled && s.Functions.UpdateEncryptedPasswordByUserID == nil {
		log.Warn("The function 'UpdateEncryptedPasswordByUserID' is undefined: the weak password hashes are not upgraded.")
	}
	if s.Config.JWT.Secret == "" {
		if needsJWT {
			return errors.New("the setting 'jwt.secret' is required by the features of the setting api_features")
		}

		router.DisableAuthentication()
	}
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
		err = httpserver.ListenAndServeTLS(address, s.Config.Server.SSL.CertPath, s.Config.Server.SSL.KeyPath, server)
	} else {
		err = httpserver.ListenAndServe(address, server)
	}
	if err != nil {
		return fmt.Errorf("the API stopped: %w", err)
	}

	return nil
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

// configureProxies sets the proxies trusted to give the address of the client, which limits the requests. No proxy is
// trusted unless the setting server.trusted_proxies lists them: Gin trusts every proxy by default, so a client could
// forge its address. It returns an error if the setting is invalid.
func (s *Server) configureProxies(server *gin.Engine) error {
	if err := server.SetTrustedProxies(s.Config.Server.TrustedProxies); err != nil {
		return fmt.Errorf("invalid setting 'server.trusted_proxies': %w", err)
	}

	return nil
}

// minimumSecretLength is the length under which the secret of the JWT can be guessed.
const minimumSecretLength = 32

// warnWeakSecret logs a warning if the secret signing the auth tokens is short.
func (s *Server) warnWeakSecret() {
	if secret := s.Config.JWT.Secret; secret != "" && len(secret) < minimumSecretLength {
		log.Warn(fmt.Sprintf("The setting 'jwt.secret' is shorter than %d characters: the auth tokens can be forged.", minimumSecretLength))
	}
}
