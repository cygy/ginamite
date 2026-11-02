package ginamite

import (
	"errors"
	"io"
	"os"
	"slices"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/timezone"
	"github.com/cygy/ginamite/web/robots"
)

// dedupSize is the number of messages the consumers remember to ignore the ones delivered twice.
const dedupSize = 10000

// Run parses the command line arguments, sets up the configuration, the translations and the services needed by the
// chosen server, configures the plugins, then runs the API, worker or web server, or the command of a plugin. It
// returns nil at once if the help is requested, and an error if the arguments or the set up are invalid. Otherwise it
// returns once the server stopped, after the consumers of the broker are stopped, the producer is flushed and the
// plugins are closed.
//
//	if err := server.Run(); err != nil {
//		log.Fatal(err)
//	}
func (s *Server) Run() error {
	f := flags.NewFlags()
	ok, err := f.Parse(s.ApplicationConfiguration.FileRequired)
	if err != nil || !ok {
		return err
	}

	return s.run(f, os.Stdout)
}

// run runs the command or the server chosen by the arguments. The output of a command is written to out.
func (s *Server) run(f *flags.Flags, out io.Writer) error {
	if handled, err := s.runCommand(f, out); handled {
		return err
	}

	if f.DLQ != "" {
		return errors.New("no plugin handles the command on the dead letters: add the plugin of the message broker")
	}

	server := s.specializedServer(f)
	if server == nil {
		return errors.New("no server selected: use one of the arguments -api, -worker or -web")
	}

	if err := s.initializeConfiguration(f); err != nil {
		return err
	}

	if f.IsAPIServer || f.IsWorkerServer {
		closeAll, err := s.connectPlugins(s.Config)
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

	defer s.startMetrics()()

	return server.Run()
}

// runCommand asks the plugins which are Commanders to run the command of the arguments, and reports whether one did.
func (s *Server) runCommand(f *flags.Flags, out io.Writer) (handled bool, err error) {
	for _, plugin := range s.plugins {
		if commander, ok := plugin.(Commander); ok {
			if handled, err = commander.Command(f, out); handled {
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
// closes them, in the reverse order. If a connection fails, it closes the connectors already connected and returns
// the error.
func (s *Server) connectPlugins(configuration *config.Configuration) (closeAll func(), err error) {
	var connected []Connector
	closeAll = func() {
		for _, c := range slices.Backward(connected) {
			c.Close()
		}
	}

	for _, plugin := range s.plugins {
		if connector, ok := plugin.(Connector); ok {
			if err := connector.Connect(configuration); err != nil {
				closeAll()
				return nil, err
			}
			connected = append(connected, connector)
		}
	}

	return closeAll, nil
}

// specializedServer returns the server chosen by the arguments, with its routes file if it uses one, or nil if none is.
// The API server wins over the worker, which wins over the web server.
func (s *Server) specializedServer(f *flags.Flags) runner {
	switch {
	case f.IsAPIServer:
		return s.API
	case f.IsWorkerServer:
		s.Worker.RoutesFile = f.RoutesFile
		return s.Worker
	case f.IsWebServer:
		s.Web.RoutesFile = f.RoutesFile
		return s.Web
	}

	return nil
}
