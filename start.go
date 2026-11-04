package ginamite

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"slices"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/timezone"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/internal/robots"
)

// dedupSize is the number of messages the consumers remember to ignore the ones delivered twice.
const dedupSize = 10000

// Run parses the command line arguments, then runs the command of a plugin if they ask for it. Otherwise it sets up
// the configuration, the translations and the services needed by the chosen server, configures the plugins, and runs
// the API, worker or web server. It returns nil at once if the help is requested, and an error if the arguments or the
// set up are invalid. Otherwise it returns once ctx is done and the server stopped, after the consumers of the broker
// are stopped, the producer is flushed and the plugins are closed. The command of a plugin stops with ctx too.
//
//	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
//	defer stop()
//
//	if err := server.Run(ctx); err != nil {
//		log.Fatal(err)
//	}
func (s *Server) Run(ctx context.Context) error {
	f := flags.NewFlags()
	if err := f.Parse(os.Args[1:], os.Stderr, s.registerFlags); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return err
	}

	if handled, err := s.runCommand(ctx, f, os.Stdout); handled {
		return err
	}

	if err := f.Validate(s.ApplicationConfiguration.FileRequired); err != nil {
		return err
	}

	return s.run(ctx, f)
}

// registerFlags lets the plugins which are Commanders add their flags.
func (s *Server) registerFlags(set *flag.FlagSet) {
	for _, plugin := range s.plugins {
		if commander, ok := plugin.(Commander); ok {
			commander.RegisterFlags(set)
		}
	}
}

// run runs the server chosen by the arguments, which are valid.
func (s *Server) run(ctx context.Context, f *flags.Flags) error {
	server := s.specializedServer(f)
	if server == nil {
		return errors.New("no server selected: use one of the arguments -api, -worker or -web")
	}

	if err := s.initializeConfiguration(f); err != nil {
		return err
	}

	if err := s.loadWebRoutes(f); err != nil {
		return err
	}

	if f.IsAPIServer || f.IsWorkerServer {
		closeAll, err := s.connectPlugins(s.Config, f.IsAPIServer)
		if err != nil {
			return err
		}
		defer closeAll()

		timezones, err := timezone.Load(s.Config.TimezonesFilePath)
		if err != nil {
			return err
		}

		s.Timezones = timezones
		s.API.Timezones = timezones
	}

	if f.IsWebServer {
		content, err := robots.Load(s.Config.RobotsFilePath)
		if err != nil {
			return err
		}

		s.Web.Robots = content
	}

	// The plugins are configured first: their functions and their deduplicator are ready when the first message comes.
	for _, plugin := range s.plugins {
		if err := plugin.Configure(s); err != nil {
			return err
		}
	}

	if s.startBroker(server.CreateConsumers()) {
		// Registered after the plugins are connected: the consumers stop, and the producer is flushed, before the
		// plugins close their connections.
		defer s.Broker.Close()
	}

	defer s.startMetrics(ctx)()

	return server.Run(ctx)
}

// runCommand asks the plugins which are Commanders to run the command of the arguments, and reports whether one did.
func (s *Server) runCommand(ctx context.Context, f *flags.Flags, out io.Writer) (handled bool, err error) {
	for _, plugin := range s.plugins {
		if commander, ok := plugin.(Commander); ok {
			if handled, err = commander.Command(ctx, f, out); handled {
				return handled, err
			}
		}
	}

	return false, nil
}

// initializeConfiguration sets up the main logger, the server configuration and the translations, which are shared
// with the specialized servers, and the application configuration.
func (s *Server) initializeConfiguration(f *flags.Flags) error {
	log.Initialize(f.ServerName, f.Environment == environment.Production)

	s.Config = config.NewConfiguration()
	if err := s.Config.Initialize(f.ServerConfigurationFile, f.Environment, f.Version); err != nil {
		return err
	}
	s.API.Config, s.Worker.Config, s.Web.Config = s.Config, s.Config, s.Config

	if s.ApplicationConfiguration.InitializeFunc != nil {
		s.ApplicationConfiguration.InitializeFunc(f.ApplicationConfigurationFile)
	}

	engine, err := localization.Load(s.Config.LocalizationPath, s.Config.SupportedLocales)
	if err != nil {
		return err
	}
	s.Localization = engine
	s.API.Localization, s.Worker.Localization, s.Web.Localization = engine, engine, engine

	return nil
}

// connectPlugins connects, with the configuration, the plugins which are Connectors, and returns the function which
// closes them, in the reverse order. The APIOnly ones are skipped if the API does not run. If a connection fails, it closes the connectors already connected and returns
// the error.
func (s *Server) connectPlugins(configuration *config.Configuration, isAPI bool) (closeAll func(), err error) {
	var connected []Connector
	closeAll = func() {
		for _, c := range slices.Backward(connected) {
			c.Close()
		}
	}

	for _, plugin := range s.plugins {
		if connector, ok := plugin.(Connector); ok {
			if _, apiOnly := plugin.(APIOnly); apiOnly && !isAPI {
				continue
			}
			if err := connector.Connect(configuration); err != nil {
				closeAll()
				return nil, err
			}
			connected = append(connected, connector)
		}
	}

	return closeAll, nil
}

// loadWebRoutes loads the web routes file of the arguments once, with the host of the website of the configuration,
// and shares the routes with the three servers. Without file, the routes are nil. It returns an error if the file
// can not be read, since the servers can not give the right URLs without it.
func (s *Server) loadWebRoutes(f *flags.Flags) error {
	s.webRoutes = nil
	if f.WebRoutesFile == "" {
		if f.IsAPIServer || f.IsWorkerServer {
			log.Warn("no web routes file (argument -web-routes): the URLs of the website are not available", "routes", "nil")
		}
	} else {
		routes, err := weburl.Load(f.WebRoutesFile, s.Config.Hosts.Web)
		if err != nil {
			return err
		}
		s.webRoutes = routes
		log.Info("web routes loaded", "file", f.WebRoutesFile, "count", len(routes.Routes))
	}
	s.API.WebRoutes, s.Worker.Routes, s.Web.Routes = s.webRoutes, s.webRoutes, s.webRoutes

	return nil
}

// specializedServer returns the server chosen by the arguments, or nil if none is.
// The API server wins over the worker, which wins over the web server. The web server also gets the routes file,
// which describes its pages.
func (s *Server) specializedServer(f *flags.Flags) runner {
	switch {
	case f.IsAPIServer:
		return s.API
	case f.IsWorkerServer:
		return s.Worker
	case f.IsWebServer:
		s.Web.RoutesFile = f.WebRoutesFile
		return s.Web
	}

	return nil
}
