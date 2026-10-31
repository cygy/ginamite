// Package web is the web server: it serves the HTML pages described by the routes file, and forwards the calls to the
// API.
package web

import (
	"fmt"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/ctxlocalization"
	"github.com/cygy/ginamite/common/localization"
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
	// Config is the configuration of the application, set by the main server before it starts.
	Config *config.Configuration
	// Localization translates the messages of the pages, set by the main server before it starts.
	Localization *localization.Engine
	// Robots is the content of the robots.txt file, loaded by the main server before it starts.
	Robots     string
	RoutesFile string // The path of the routes file.
	Functions  Handlers

	customRoutes []route.ConfigureCustomRoutes
	apiClient    *api.Client         // The client of the API, created when the server is set up.
	pages        *html.Configuration // The pages, created when the server is set up.
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
		Config:       config.NewConfiguration(),
		Localization: localization.NewEngine(),
		customRoutes: []route.ConfigureCustomRoutes{},
	}
}

// AddCustomRoutes adds some custom routes, which are loaded for each locale when the server starts.
func (s *Server) AddCustomRoutes(configuration route.ConfigureCustomRoutes) {
	s.customRoutes = append(s.customRoutes, configuration)
}

// CreateConsumers returns the Kafka consumers of the web server, which has none.
func (s *Server) CreateConsumers() []broker.MessageConsumer {
	return []broker.MessageConsumer{}
}

// Start sets up and starts the web server, which stops on error.
func (s *Server) Start() {
	s.warnUndefinedFunctions()

	engine := s.newEngine()
	s.setUp(engine)

	address := fmt.Sprintf(":%d", s.Config.Server.Port)

	var err error
	if s.Config.Server.SSL.Enabled {
		err = listenAndServeTLS(address, s.Config.Server.SSL.CertPath, s.Config.Server.SSL.KeyPath, engine)
	} else {
		err = listenAndServe(address, engine)
	}

	if err != nil {
		log.WithField("error", err).Error("the server stopped")
	}
}

// newEngine returns the Gin engine, in debug mode if enabled by the configuration. The requests are logged by the
// request logger in production, by the default logger else, and the responses are gzipped if enabled.
func (s *Server) newEngine() *gin.Engine {
	if s.Config.IsDebugModeEnabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	var engine *gin.Engine
	if s.Config.IsProductionEnvironment() {
		engine = gin.New()
		engine.Use(gin.Recovery())
		engine.Use(log.InjectRequestLogger(false))
	} else {
		engine = gin.Default()
	}

	if s.Config.Server.GZip.Enabled {
		engine.Use(gzip.Gzip(s.Config.Server.GZip.Level))
	}

	return engine
}

// setUp initializes the API client, the HTML templates and pages, and loads the routes into the engine. The pages are
// saved in the context of the requests, with the translation engine.
func (s *Server) setUp(engine *gin.Engine) {
	client := api.NewClient(s.Config.API.Host, s.Config.API.Version, s.Config.API.Port, s.Config.API.TimeOut, s.Config.API.RetryCount, s.Config.IsDebugModeEnabled())
	s.apiClient = client
	client.AddResponseHandlers(append([]api.ResponseHandlerFunc{
		response.HandleInternalServerError(),
		response.HandleInvalidAuthenticationError(),
		response.HandleLatestTermsVersionMustBeAcceptedError(),
	}, s.Functions.APIMiddlewares...)...)

	pages := html.Load(s.RoutesFile, html.Settings{
		WebHost:         s.Config.Hosts.Web,
		StaticHost:      s.Config.Hosts.Static,
		APIHost:         s.Config.Hosts.API + "/" + s.Config.APIVersion,
		Timezone:        s.Config.DefaultTimezone,
		Locales:         s.Config.SupportedLocales,
		FallbackLocales: s.Config.FallbackLocales,
		LocaleFormats:   s.Config.Locales,
		Facebook: html.FacebookConfiguration{
			Enabled:    s.Config.Facebook.Enabled,
			AppID:      s.Config.Facebook.AppID,
			APIVersion: s.Config.Facebook.APIVersion,
		},
		Google: html.GoogleConfiguration{
			Enabled:  s.Config.Google.Enabled,
			ClientID: s.Config.Google.ClientID,
		},
		SocialNetworks: html.SocialNetworksConfiguration{
			Facebook: s.Config.SocialNetworks.Facebook,
			Twitter:  s.Config.SocialNetworks.Twitter,
			Google:   s.Config.SocialNetworks.Google,
		},
		BetaVersion:  !s.Config.IsFinalVersion(),
		Production:   s.Config.IsProductionEnvironment(),
		Localization: s.Localization,
		Templates:    htmlTemplate.LoadTemplates(s.Config.TemplatesPath, engine),
	})
	s.pages = pages
	engine.Use(ctxconfig.Inject(s.Config), ctxlocalization.Inject(s.Localization), html.Inject(pages))

	router := route.NewRouter(engine, pages, client,
		s.Config.SupportedLocales, s.Config.DefaultLocale,
		s.Config.Cookie.Name, s.Config.JWT.Secret,
	)
	router.Robots = s.Robots
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
