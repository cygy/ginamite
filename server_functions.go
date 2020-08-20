package ginamite

import (
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	mongo "github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/timezone"
	"github.com/cygy/ginamite/web/robots"
)

// AddPlugin : adds a plugin to the list.
func (s *Server) AddPlugin(p Plugin) {
	s.plugins = append(s.plugins, p)
}

// Start : starts the server.
func (s *Server) Start() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())

	// Parse the command arguments.
	flags := flags.NewFlags()
	if !flags.Parse(s.ApplicationConfiguration.FileRequired) {
		os.Exit(0)
		return
	}

	// Initialize the main logger.
	log.Initialize(flags.ServerName, (flags.Environment == environment.Production))

	// Set up the server configuration.
	config.Initialize(flags.ServerConfigurationFile, flags.Environment, flags.Version)

	// Set up the application configuration.
	if s.ApplicationConfiguration.InitializeFunc != nil {
		s.ApplicationConfiguration.InitializeFunc(flags.ApplicationConfigurationFile)
	}

	// Set up the translation engines.
	localization.Initialize(config.Main.LocalizationPath, config.Main.SupportedLocales)

	// API/Worker relative settings.
	if flags.IsAPIServer || flags.IsWorkerServer {
		// Set up the database connection.
		if config.Main.IsMongoDBDatabase() {
			mongo.Initialize(config.Main.Database.Host+":"+strconv.Itoa(config.Main.Database.Port),
				config.Main.Database.Database,
				config.Main.Database.Username,
				config.Main.Database.Password,
				time.Duration(config.Main.Database.TimeOut))
			defer mongo.Close()
		}

		// Define the timezones.
		timezone.Initialize(config.Main.TimezonesFilePath)
	}

	// Web relative settings.
	if flags.IsWebServer {
		// Load the robots.txt file.
		robots.Initialize(config.Main.RobotsFilePath)
	}

	// Create the specialized server.
	var specializedServer SpecializedServer

	if flags.IsAPIServer {
		s.API.RoutesFile = flags.RoutesFile
		specializedServer = s.API
	} else if flags.IsWorkerServer {
		s.Worker.RoutesFile = flags.RoutesFile
		specializedServer = s.Worker
	} else if flags.IsWebServer {
		s.Web.RoutesFile = flags.RoutesFile
		specializedServer = s.Web
	}

	// Initialize the kafka messages (consumer/producer).
	if kafkaConf, ok := config.Main.KafkaConfiguration(); ok {
		consumers := specializedServer.CreateKafkaConsumers()
		kafka.Initialize(kafkaConf.Host, kafkaConf.Port, config.Main.KafkaTopicsPrefix, true, consumers)
	}

	// Loads the plugins.
	for _, plugin := range s.plugins {
		plugin.Configure(s)
	}

	// Launch the specialized server.
	specializedServer.Start()
}

// AddSupportedNotification : adds some supported notification types.
func (a ApplicationConfiguration) AddSupportedNotification(notification string, isDefault bool) {
	notifications.AddNotification(notification, isDefault)
}

// DisableNotification : disables a supported notification.
func (a ApplicationConfiguration) DisableNotification(notification string) bool {
	return notifications.DisableNotification(notification)
}

// EnableNotification : disables a supported notification.
func (a ApplicationConfiguration) EnableNotification(notification string) bool {
	return notifications.EnableNotification(notification)
}

// SetNotificationDefault : sets a supported notification as default or not.
func (a ApplicationConfiguration) SetNotificationDefault(notification string, isDefault bool) bool {
	return notifications.SetNotificationDefault(notification, isDefault)
}
