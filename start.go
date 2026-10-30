package ginamite

import (
	"os"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/timezone"
	"github.com/cygy/ginamite/web/robots"
)

// Start parses the command line arguments, sets up the configuration, the translations and the services needed by the
// chosen server, configures the plugins, then starts the API, worker or web server. It exits the process if the
// help is requested or if the set up fails.
func (s *Server) Start() {
	f := flags.NewFlags()
	if !f.Parse(s.ApplicationConfiguration.FileRequired) {
		os.Exit(0)
		return
	}

	s.initializeConfiguration(f)

	if f.IsAPIServer || f.IsWorkerServer {
		defer s.connectPlugins()()

		if err := timezone.Initialize(config.Main.TimezonesFilePath); err != nil {
			log.WithField("error", err).Fatal("unable to initialize the timezones")
		}
	}

	if f.IsWebServer {
		if err := robots.Initialize(config.Main.RobotsFilePath); err != nil {
			log.WithField("error", err).Fatal("unable to initialize the robots.txt file")
		}
	}

	specializedServer := s.specializedServer(f)
	if specializedServer == nil {
		log.Fatal("no server selected: use one of the arguments -api, -worker or -web")
		return
	}

	if kafkaConf, ok := config.Main.KafkaConfiguration(); ok {
		kafka.Initialize(kafkaConf.Host, kafkaConf.Port, config.Main.KafkaTopicsPrefix, true, specializedServer.CreateKafkaConsumers())
	}

	for _, plugin := range s.plugins {
		plugin.Configure(s)
	}

	specializedServer.Start()
}

// initializeConfiguration sets up the main logger, the server and application configurations and the translations.
func (s *Server) initializeConfiguration(f *flags.Flags) {
	log.Initialize(f.ServerName, f.Environment == environment.Production)

	if err := config.Initialize(f.ServerConfigurationFile, f.Environment, f.Version); err != nil {
		log.WithField("error", err).Fatal("unable to initialize the server configuration")
	}

	if s.ApplicationConfiguration.InitializeFunc != nil {
		s.ApplicationConfiguration.InitializeFunc(f.ApplicationConfigurationFile)
	}

	localization.Initialize(config.Main.LocalizationPath, config.Main.SupportedLocales)
}

// connectPlugins connects the plugins which are Connectors, and returns the function which closes them, in the reverse
// order.
func (s *Server) connectPlugins() (closeAll func()) {
	var connected []Connector
	for _, plugin := range s.plugins {
		if connector, ok := plugin.(Connector); ok {
			connector.Connect()
			connected = append(connected, connector)
		}
	}

	return func() {
		for i := len(connected) - 1; i >= 0; i-- {
			connected[i].Close()
		}
	}
}

// specializedServer returns the server chosen by the arguments, with its routes file, or nil if none is.
// The API server wins over the worker, which wins over the web server.
func (s *Server) specializedServer(f *flags.Flags) SpecializedServer {
	switch {
	case f.IsAPIServer:
		s.API.RoutesFile = f.RoutesFile
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
