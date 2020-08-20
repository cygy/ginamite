package web

import (
	"fmt"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"
	htmlTemplate "github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/response"
	"github.com/cygy/ginamite/web/route"

	"github.com/fvbock/endless"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// CreateKafkaConsumers return an array of kafka consumers.
func (s *Server) CreateKafkaConsumers() []kafka.MessageConsumer {
	consumers := []kafka.MessageConsumer{}

	return consumers
}

// Start start the worker.
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
		server.Use(log.InjectRequestLogger(false))
	} else {
		server = gin.Default()
	}

	// Enable server settings
	if config.Main.Server.GZip.Enabled {
		server.Use(gzip.Gzip(config.Main.Server.GZip.Level))
	}

	// Set up the API client.
	api.Initialize(config.Main.API.Host, config.Main.API.Version, config.Main.API.Port, config.Main.API.TimeOut, config.Main.API.RetryCount, config.Main.IsDebugModeEnabled())

	handlers := []api.ResponseHandlerFunc{
		response.HandleInternalServerError(),
		response.HandleInvalidAuthenticationError(),
		response.HandleLatestTermsVersionMustBeAcceptedError(),
	}
	handlers = append(handlers, s.Functions.APIMiddlewares...)
	api.AddResponseHandlers(handlers...)

	// Load the HTML templates.
	htmlTemplate.LoadTemplates(config.Main.TemplatesPath, server)

	// Initialize the configuration of the HTML routes.
	html.Initialize(s.RoutesFile,
		config.Main.Hosts.Web, config.Main.Hosts.Static, config.Main.Hosts.API+"/"+config.Main.APIVersion,
		config.Main.DefaultTimezone, config.Main.SupportedLocales, config.Main.FallbackLocales,
		config.Main.Cookie.Name, config.Main.JWT.Secret,
		html.FacebookConfiguration{
			Enabled:    config.Main.Facebook.Enabled,
			AppID:      config.Main.Facebook.AppID,
			APIVersion: config.Main.Facebook.APIVersion,
		},
		html.GoogleConfiguration{
			Enabled:  config.Main.Google.Enabled,
			ClientID: config.Main.Google.ClientID,
		},
		html.SocialNetworksConfiguration{
			Facebook: config.Main.SocialNetworks.Facebook,
			Twitter:  config.Main.SocialNetworks.Twitter,
			Google:   config.Main.SocialNetworks.Google,
		},
		!config.Main.IsFinalVersion(),
		config.Main.IsProductionEnvironment(),
	)

	// Set up the routes.
	router := route.NewRouter(server,
		config.Main.SupportedLocales, config.Main.DefaultLocale,
		config.Main.Cookie.Name, config.Main.JWT.Secret,
	)
	if s.Functions.RouterConfigureDefaultRoutes != nil {
		s.Functions.RouterConfigureDefaultRoutes(&router.DefaultRoutes)
	}
	router.LoadDefaultRoutes()
	for _, customRoute := range s.customRoutes {
		router.LoadCustomRoutes(customRoute)
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
}
