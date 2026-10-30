// Package web is the web server: it serves the HTML pages described by the routes file, and forwards the calls to the
// API.
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

// listenAndServe and listenAndServeTLS start the HTTP server; the tests replace them.
var (
	listenAndServe    = endless.ListenAndServe
	listenAndServeTLS = endless.ListenAndServeTLS
)

// Server is a web server.
type Server struct {
	RoutesFile string // The path of the routes file.
	Functions  Handlers

	customRoutes []route.ConfigureCustomRoutes
}

// Handlers are the functions of the application used by the web server.
type Handlers struct {
	RecurringTasks               func()
	APIMiddlewares               []api.ResponseHandlerFunc // The response handlers of the API calls, after the built-in ones.
	RouterConfigureDefaultRoutes route.ConfigureDefaultRoutes
}

// NewServer returns a web server without any route.
func NewServer() *Server {
	return &Server{
		customRoutes: []route.ConfigureCustomRoutes{},
	}
}

// AddCustomRoutes adds some custom routes, which are loaded for each locale when the server starts.
func (s *Server) AddCustomRoutes(configuration route.ConfigureCustomRoutes) {
	s.customRoutes = append(s.customRoutes, configuration)
}

// CreateKafkaConsumers returns the Kafka consumers of the web server, which has none.
func (s *Server) CreateKafkaConsumers() []kafka.MessageConsumer {
	return []kafka.MessageConsumer{}
}

// Start sets up and starts the web server, which stops on error.
func (s *Server) Start() {
	s.warnUndefinedFunctions()

	engine := newEngine()
	s.setUp(engine)

	address := fmt.Sprintf(":%d", config.Main.Server.Port)

	var err error
	if config.Main.Server.SSL.Enabled {
		err = listenAndServeTLS(address, config.Main.Server.SSL.CertPath, config.Main.Server.SSL.KeyPath, engine)
	} else {
		err = listenAndServe(address, engine)
	}

	if err != nil {
		log.WithField("error", err).Error("the server stopped")
	}
}

// newEngine returns the Gin engine, in debug mode if enabled by the configuration. The requests are logged by the
// request logger in production, by the default logger else, and the responses are gzipped if enabled.
func newEngine() *gin.Engine {
	if config.Main.IsDebugModeEnabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	var engine *gin.Engine
	if config.Main.IsProductionEnvironment() {
		engine = gin.New()
		engine.Use(gin.Recovery())
		engine.Use(log.InjectRequestLogger(false))
	} else {
		engine = gin.Default()
	}

	if config.Main.Server.GZip.Enabled {
		engine.Use(gzip.Gzip(config.Main.Server.GZip.Level))
	}

	return engine
}

// setUp initializes the API client, the HTML templates and pages, and loads the routes into the engine.
func (s *Server) setUp(engine *gin.Engine) {
	api.Initialize(config.Main.API.Host, config.Main.API.Version, config.Main.API.Port, config.Main.API.TimeOut, config.Main.API.RetryCount, config.Main.IsDebugModeEnabled())
	api.AddResponseHandlers(append([]api.ResponseHandlerFunc{
		response.HandleInternalServerError(),
		response.HandleInvalidAuthenticationError(),
		response.HandleLatestTermsVersionMustBeAcceptedError(),
	}, s.Functions.APIMiddlewares...)...)

	htmlTemplate.LoadTemplates(config.Main.TemplatesPath, engine)

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

	router := route.NewRouter(engine,
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
}

// warnUndefinedFunctions logs a warning for each function of the application which is not defined.
func (s *Server) warnUndefinedFunctions() {
	if s.Functions.RecurringTasks == nil {
		log.Warn("The function 'RecurringTasks' is undefined.")
	}
}
