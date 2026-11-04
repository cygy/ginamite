// Package ginamite is a framework to build an application made of an API server, a worker server and a web server.
//
// A Server holds the three specialized servers and the settings of the application. The command line arguments choose
// the one which runs, see the package flags.
//
// The storage is provided by plugins, see Plugin and Connector: the core does not depend on a database.
package ginamite

import (
	"context"
	"flag"
	"io"
	"time"

	"github.com/cygy/ginamite/api"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/timezone"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/web"
	"github.com/cygy/ginamite/worker"
)

// Server is the main struct of the application, containing all the settings.
type Server struct {
	// Config is the configuration of the application, loaded from the configuration file by Run and shared with the
	// specialized servers and the plugins.
	Config                   *config.Configuration
	Broker                   broker.Broker           // The message broker, none by default: a plugin sets it, such as the one of Kafka, see UseBroker. Started by Run.
	Deduplicator             broker.Deduplicator     // Remembers the messages handled, to ignore the ones delivered twice. In memory by default; a plugin may replace it.
	Queue                    *queue.Queue            // Sends the messages to the other servers through the broker.
	Localization             *localization.Engine    // The translations, loaded by Run.
	Timezones                *timezone.Configuration // The timezones of the countries, loaded by Run for the API and the worker.
	ApplicationConfiguration ApplicationConfiguration
	Worker                   *worker.Server
	API                      *api.Server
	Web                      *web.Server

	// private properties
	plugins   []Plugin
	webRoutes *weburl.Configuration
}

// runner is a specialized server: API, worker or web.
type runner interface {
	CreateConsumers() []broker.MessageConsumer
	Run(ctx context.Context) error
}

// UseBroker sets the message broker, and the queue sending the messages through it, which the API and the worker use.
// A plugin calls it when it is configured, or the application before Run, which starts the broker with the consumers
// of the server. Without a broker, the messages are not sent and an error is logged for each.
func (s *Server) UseBroker(b broker.Broker) {
	s.Broker = b
	s.Queue = queue.New(b)
	s.API.Queue, s.Worker.Queue = s.Queue, s.Queue
}

// Plugin is a plugin which configures a server, such as the functions of the API and of the workers.
//
// Run calls Configure once the configuration is loaded, before it starts the server, and stops with the error if
// Configure returns one.
type Plugin interface {
	Configure(s *Server) error
}

// Commander is a Plugin which runs a command given by the command line arguments, instead of a server.
//
// Run calls RegisterFlags before it parses the command line arguments, so the plugin adds the flags of its command,
// then Command with the context of Run and the arguments, before it checks that they describe a server and sets it up. A command handled by a
// plugin is the last thing Run does. The plugins are asked in the order they were added, until one handles the
// command.
type Commander interface {
	Plugin
	RegisterFlags(set *flag.FlagSet)
	Command(ctx context.Context, f *flags.Flags, out io.Writer) (handled bool, err error)
}

// Connector is a Plugin which holds a connection to a database, or to any other service the API and the worker need.
//
// Run calls Connect with the configuration, before it configures the plugins, only if the API or the worker server
// runs, then Close when the server stops. Run stops with the error if Connect returns one, after closing the
// connectors already connected.
type Connector interface {
	Plugin
	Connect(configuration *config.Configuration) error
	Close()
}

// APIOnly is a Connector which the API alone needs: Run does not connect it when only the worker runs.
type APIOnly interface {
	Connector
	APIOnly()
}

// NewServer returns a Server whose API functions are placeholders which accept nothing: a plugin or the application
// is expected to replace them.
func NewServer() *Server {
	server := &Server{
		Config:       config.NewConfiguration(),
		Deduplicator: broker.NewMemoryDeduplicator(dedupSize),
		Queue:        queue.New(nil),
		Localization: localization.NewEngine(),
		Timezones:    timezone.NewConfiguration(),
		plugins:      []Plugin{},
		Web:          web.NewServer(),
		Worker:       worker.NewServer(),
		API:          api.NewServer(),
	}
	server.API.Queue, server.Worker.Queue = server.Queue, server.Queue

	server.ApplicationConfiguration.Notifications = notifications.NewRegistry()
	server.API.Notifications = server.ApplicationConfiguration.Notifications
	server.ApplicationConfiguration.FileRequired = false

	server.API.Hooks.NewMessageFromQueueCache = func(string) bool {
		return true
	}
	server.API.Hooks.BuildTokenFromID = func(context.Context, string) (*authentication.Token, error) {
		return &authentication.Token{}, nil
	}
	server.API.Hooks.ExtraPropertiesForTokenWithID = func(context.Context, string) (map[string]string, error) {
		return map[string]string{}, nil
	}
	server.API.Hooks.ExtendTokenExpirationDateFromID = func(context.Context, string, time.Duration) error {
		return nil
	}
	server.API.Hooks.MiddlewareIsValidAuthToken = func(context.Context, string) (bool, error) {
		return false, nil
	}
	server.API.Hooks.MiddlewareGetUserAbilities = func(context.Context, string) (bool, []string, error) {
		return false, []string{}, nil
	}

	return server
}

// WebRoutes returns the routes of the website, loaded by Run from the file given by the argument -web-routes and
// shared with the three servers. It is nil if no file is given, which the API and the worker allow: a nil
// *weburl.Configuration is safe to use, its URLs are empty. The plugins read it in Configure, the application in the
// functions it gives to the servers, after Run loaded it.
func (s *Server) WebRoutes() *weburl.Configuration {
	return s.webRoutes
}

// AddPlugin adds a plugin, which is configured by Run.
func (s *Server) AddPlugin(p Plugin) {
	s.plugins = append(s.plugins, p)
}
