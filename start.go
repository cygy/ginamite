package ginamite

import (
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

// Start parses the command line arguments, sets up the configuration, the translations and the services needed by the
// chosen server, configures the plugins, then starts the API, worker or web server. It exits the process if the
// help is requested or if the set up fails. It returns once the server stopped, after the consumers of Kafka are
// stopped, the producer is flushed and the plugins are closed.
func (s *Server) Start() {
	f := flags.NewFlags()
	if !f.Parse(s.ApplicationConfiguration.FileRequired) {
		os.Exit(0)
		return
	}

	if handled, err := s.runCommand(f, os.Stdout); handled {
		if err != nil {
			log.WithField("error", err).Fatal("the command failed")
		}

		return
	}

	if f.DLQ != "" {
		log.Fatal("no plugin handles the command on the dead letters: add the plugin of the message broker")
		return
	}

	s.initializeConfiguration(f)

	if f.IsAPIServer || f.IsWorkerServer {
		defer s.connectPlugins(s.Config)()

		timezones, err := timezone.Load(s.Config.TimezonesFilePath)
		if err != nil {
			log.WithField("error", err).Fatal("unable to initialize the timezones")
			return
		}

		s.Timezones = timezones
		s.API.Timezones = timezones
	}

	if f.IsWebServer {
		content, err := robots.Load(s.Config.RobotsFilePath)
		if err != nil {
			log.WithField("error", err).Fatal("unable to initialize the robots.txt file")
			return
		}

		s.Web.Robots = content
	}

	specializedServer := s.specializedServer(f)
	if specializedServer == nil {
		log.Fatal("no server selected: use one of the arguments -api, -worker or -web")
		return
	}

	// The plugins are configured first: their functions and their deduplicator are ready when the first message comes.
	for _, plugin := range s.plugins {
		plugin.Configure(s)
	}

	if s.startBroker(specializedServer.CreateConsumers()) {
		// Registered after the plugins are connected: the consumers stop, and the producer is flushed, before the
		// plugins close their connections.
		defer s.Broker.Close()
	}

	defer s.startMetrics()()

	specializedServer.Start()
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
func (s *Server) initializeConfiguration(f *flags.Flags) {
	log.Initialize(f.ServerName, f.Environment == environment.Production)

	s.Config = config.NewConfiguration()
	if err := s.Config.Initialize(f.ServerConfigurationFile, f.Environment, f.Version); err != nil {
		log.WithField("error", err).Fatal("unable to initialize the server configuration")
	}
	s.API.Config, s.Worker.Config, s.Web.Config = s.Config, s.Config, s.Config

	if s.ApplicationConfiguration.InitializeFunc != nil {
		s.ApplicationConfiguration.InitializeFunc(f.ApplicationConfigurationFile)
	}

	s.Localization = localization.Load(s.Config.LocalizationPath, s.Config.SupportedLocales)
	s.API.Localization, s.Worker.Localization, s.Web.Localization = s.Localization, s.Localization, s.Localization
}

// connectPlugins connects, with the configuration, the plugins which are Connectors, and returns the function which closes them, in the reverse
// order.
func (s *Server) connectPlugins(configuration *config.Configuration) (closeAll func()) {
	var connected []Connector
	for _, plugin := range s.plugins {
		if connector, ok := plugin.(Connector); ok {
			connector.Connect(configuration)
			connected = append(connected, connector)
		}
	}

	return func() {
		for _, c := range slices.Backward(connected) {
			c.Close()
		}
	}
}

// specializedServer returns the server chosen by the arguments, with its routes file if it uses one, or nil if none is.
// The API server wins over the worker, which wins over the web server.
func (s *Server) specializedServer(f *flags.Flags) SpecializedServer {
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
