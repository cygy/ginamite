// Package web is the web server: it serves the HTML pages described by the routes file, and forwards the calls to the
// API.
package web

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/request"
	htmlTemplate "github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/internal/httpserver"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/response"
	"github.com/cygy/ginamite/web/route"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// listenAndServe and listenAndServeTLS start the HTTP server; the tests replace them.
var (
	listenAndServe    = httpserver.ListenAndServe
	listenAndServeTLS = httpserver.ListenAndServeTLS
)

// Server is a web server.
type Server struct {
	// Config is the configuration of the application, set by the main server before it starts.
	Config *config.Configuration
	// Localization translates the messages of the pages, set by the main server before it starts.
	Localization *localization.Engine
	// Robots is the content of the robots.txt file, loaded by the main server before it starts.
	Robots string
	// RoutesFile is the path of the routes file, which describes the pages. Set by the main server before it starts.
	RoutesFile string
	// Routes gives the URLs of the routes of the file, set by the main server before it starts; the handlers read
	// them with weburl.Of.
	Routes    *weburl.Configuration
	Functions Handlers

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

// Run sets up the web server, then serves it until ctx is done. It returns an error if the set up fails or if the
// server stops for another reason.
func (s *Server) Run(ctx context.Context) error {
	engine, err := s.newEngine()
	if err != nil {
		return err
	}
	if err := s.setUp(engine); err != nil {
		return err
	}

	address := fmt.Sprintf(":%d", s.Config.Server.Port)

	if s.Config.Server.SSL.Enabled {
		err = listenAndServeTLS(ctx, address, s.Config.Server.SSL.CertPath, s.Config.Server.SSL.KeyPath, engine)
	} else {
		err = listenAndServe(ctx, address, engine)
	}

	if err != nil {
		return fmt.Errorf("the web server stopped: %w", err)
	}

	return nil
}

// newEngine returns the Gin engine, in debug mode if enabled by the configuration. The requests are logged by the
// request logger in production, by the default logger else, and the responses are gzipped if enabled. Only the proxies
// of the setting server.trusted_proxies give the address of the client, which the web server sends to the API. The
// cross-origin requests with an unsafe method are rejected, except from hosts.web, and every response has the security
// headers, with HSTS in production. It returns an error if a setting is invalid.
func (s *Server) newEngine() (*gin.Engine, error) {
	if s.Config.IsDebugModeEnabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// The body is limited before the logger reads it.
	engine := gin.New()
	engine.Use(gin.Recovery(), request.LimitBodySize(s.Config.Server.MaxBodySize), request.SecurityHeaders(s.Config.IsProductionEnvironment()))
	if s.Config.IsProductionEnvironment() {
		engine.Use(log.InjectRequestLogger(false))
	} else {
		engine.Use(gin.Logger())
	}

	if err := request.ConfigureClientIP(engine, s.Config.Server.TrustedProxies); err != nil {
		return nil, fmt.Errorf("invalid setting 'server.trusted_proxies': %w", err)
	}

	// The forms and the calls forwarded to the API are authenticated by the cookie: the cross-origin ones are rejected.
	crossOrigin, err := request.CrossOriginProtection(s.Config.Hosts.Web)
	if err != nil {
		return nil, fmt.Errorf("invalid setting 'hosts.web': %w", err)
	}
	engine.Use(crossOrigin)

	if s.Config.Server.GZip.Enabled {
		engine.Use(gzip.Gzip(s.Config.Server.GZip.Level))
	}

	return engine, nil
}

// setUp initializes the API client, the HTML templates and pages, and loads the routes into the engine. The pages are
// saved in the context of the requests, with the translation engine. It returns an error if the templates or the
// routes file can not be loaded.
func (s *Server) setUp(engine *gin.Engine) error {
	client := api.NewClient(s.Config.API.Host, s.Config.API.Version, s.Config.API.Port, s.Config.API.TimeOut, s.Config.API.RetryCount, s.Config.IsDebugModeEnabled())
	s.apiClient = client
	client.AddResponseHandlers(append([]api.ResponseHandlerFunc{
		response.HandleInternalServerError(),
		response.HandleInvalidAuthenticationError(),
		response.HandleLatestTermsVersionMustBeAcceptedError(),
	}, s.Functions.APIMiddlewares...)...)

	templates, err := htmlTemplate.LoadTemplates(s.Config.TemplatesPath, engine)
	if err != nil {
		return err
	}

	pages, err := html.Load(s.RoutesFile, html.Settings{
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
		Templates:    templates,
	})
	if err != nil {
		return err
	}
	s.pages = pages
	engine.Use(ctxkey.Config.Inject(s.Config), ctxkey.Localization.Inject(s.Localization), html.Inject(pages), weburl.Inject(s.Routes))

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

	return nil
}
