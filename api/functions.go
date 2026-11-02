// Package api wires the HTTP API server: routes, middlewares and the functions provided by the application.
package api

import (
	"errors"
	"fmt"
	"slices"
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
	commonRequest "github.com/cygy/ginamite/common/request"
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
				if s.Hooks.NewMessageFromQueueCache == nil || !s.Hooks.NewMessageFromQueueCache(msgType) {
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

	// Load the middlewares.
	if s.Requests != nil {
		server.Use(s.Requests.Middleware())
	}
	server.Use(ctxkey.Config.Inject(s.Config), ctxkey.Queue.Inject(s.Queue), ctxkey.Localization.Inject(s.Localization), ctxkey.Timezones.Inject(s.Timezones), ctxkey.Notifications.Inject(s.Notifications))
	server.Use(s.Middlewares...)
	server.Use(s.idValidator())
	server.Use(ctxAuthToken.InjectRevoked(s.RevokedTokens.IsRevoked))
	server.Use(ctxAuthToken.InjectHooks(authentication.Hooks{
		BuildTokenFromID:                s.Hooks.BuildTokenFromID,
		ExtraPropertiesForTokenWithID:   s.Hooks.ExtraPropertiesForTokenWithID,
		ExtendTokenExpirationDateFromID: s.Hooks.ExtendTokenExpirationDateFromID,
	}))
	server.Use(
		account.Inject(&s.Account),
		user.Inject(s.Users),
		contact.Inject(s.Contacts),
		terms.Inject(s.Terms),
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
		s.Hooks.MiddlewareIsValidAuthToken,
		s.Hooks.MiddlewareGetUserAndLocaleFromAuthToken,
		s.Hooks.MiddlewareGetUserAbilities,
	)
	needsJWT, err := s.selectRoutes(&router.DefaultRoutes, s.Config.APIFeatures)
	if err != nil {
		return fmt.Errorf("unable to load the routes of the API: %w", err)
	}
	if err := s.checkProviders(); err != nil {
		return err
	}
	if s.Config.JWT.Secret == "" {
		if needsJWT {
			return errors.New("the setting 'jwt.secret' is required by the features of the setting api_features")
		}

		router.DisableAuthentication()
	}
	if s.Hooks.RouterConfigureDefaultRoutes != nil {
		s.Hooks.RouterConfigureDefaultRoutes(&router.DefaultRoutes)
	}
	router.Notifications = s.Notifications
	router.Limits = route.RateLimits{
		Authentication: request.RateLimitFromConfiguration("authentication", s.Config.RateLimits.Authentication, s.Hooks.NewRateLimiter),
		Registration:   request.RateLimitFromConfiguration("registration", s.Config.RateLimits.Registration, s.Hooks.NewRateLimiter),
		ForgotPassword: request.RateLimitFromConfiguration("forgot_password", s.Config.RateLimits.ForgotPassword, s.Hooks.NewRateLimiter),
	}
	router.LoadDefaultRoutes()
	for _, customRoute := range s.customRoutes {
		customRoute(router.Group, router.Handlers, router.ProductionEnvironment, router.Rights)
	}

	// Start the built-in recurring tasks.
	s.RevokedTokens.StartDeletingExpired(time.Duration(s.Config.ExpiredAuthTokens.FlushCacheInterval))

	// Start the custom recurring tasks.
	if s.Hooks.RecurringTasks != nil {
		s.Hooks.RecurringTasks()
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

// idValidator returns the middleware making the requests use the IsValidID function of the server, if any, to validate
// the identifiers. It does nothing if the function is undefined.
func (s *Server) idValidator() gin.HandlerFunc {
	if s.Hooks.IsValidID == nil {
		return func(*gin.Context) {}
	}

	return request.InjectIDValidator(s.Hooks.IsValidID)
}

// configureProxies sets the proxies trusted to give the address of the client, which limits the requests and is saved
// with the auth tokens. No proxy is trusted unless the setting server.trusted_proxies lists them: Gin trusts every proxy
// by default, so a client could forge its address. List the web servers too: they give the address of the user in the
// header X-User-IP. It returns an error if the setting is invalid.
func (s *Server) configureProxies(server *gin.Engine) error {
	if err := commonRequest.ConfigureClientIP(server, s.Config.Server.TrustedProxies); err != nil {
		return fmt.Errorf("invalid setting 'server.trusted_proxies': %w", err)
	}

	return nil
}

// checkProviders returns an error if the features facebook or google are enabled without the settings which verify the
// tokens of their users.
func (s *Server) checkProviders() error {
	if slices.Contains(s.Config.APIFeatures, FeatureFacebook) && (s.Config.Facebook.AppID == "" || s.Config.Facebook.AppSecret == "") {
		return errors.New("the feature facebook needs the settings 'facebook.app_id' and 'facebook.app_secret'")
	}
	if slices.Contains(s.Config.APIFeatures, FeatureGoogle) && s.Config.Google.ClientID == "" {
		return errors.New("the feature google needs the setting 'google.client_id'")
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
