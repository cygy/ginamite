// Package ginamite is a framework to build an application made of an API server, a worker server and a web server.
//
// A Server holds the three specialized servers and the settings of the application. The command line arguments choose
// the one which runs, see the package flags.
//
// The storage is provided by plugins, see Plugin and Connector: the core does not depend on a database.
package ginamite

import (
	"context"
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
	"github.com/cygy/ginamite/web"
	"github.com/cygy/ginamite/worker"

	"github.com/gin-gonic/gin"
)

// Server is the main struct of the application, containing all the settings.
type Server struct {
	// Config is the configuration of the application, loaded from the configuration file by Start and shared with the
	// specialized servers and the plugins.
	Config                   *config.Configuration
	Broker                   broker.Broker           // The message broker, none by default: a plugin sets it, such as the one of Kafka, see UseBroker. Started by Start.
	Deduplicator             broker.Deduplicator     // Remembers the messages handled, to ignore the ones delivered twice. In memory by default; a plugin may replace it.
	Queue                    *queue.Queue            // Sends the messages to the other servers through the broker.
	Localization             *localization.Engine    // The translations, loaded by Start.
	Timezones                *timezone.Configuration // The timezones of the countries, loaded by Start for the API and the worker.
	ApplicationConfiguration ApplicationConfiguration
	Worker                   *worker.Server
	API                      *api.Server
	Web                      *web.Server

	// private properties
	plugins []Plugin
}

// runner is a specialized server: API, worker or web.
type runner interface {
	CreateConsumers() []broker.MessageConsumer
	Start()
}

// UseBroker sets the message broker, and the queue sending the messages through it, which the API and the worker use.
// A plugin calls it when it is configured, or the application before Start, which starts the broker with the consumers
// of the server. Without a broker, the messages are not sent and an error is logged for each.
func (s *Server) UseBroker(b broker.Broker) {
	s.Broker = b
	s.Queue = queue.New(b)
	s.API.Queue, s.Worker.Queue = s.Queue, s.Queue
}

// Plugin is a plugin which configures a server, such as the functions of the API and of the workers.
type Plugin interface {
	Configure(s *Server)
}

// Commander is a Plugin which runs a command given by the command line arguments, instead of a server.
//
// Start calls Command with the arguments, before it sets up a server. A command handled by a plugin is the last thing
// Start does. The plugins are asked in the order they were added, until one handles the command.
type Commander interface {
	Plugin
	Command(f *flags.Flags, out io.Writer) (handled bool, err error)
}

// Connector is a Plugin which holds a connection to a database, or to any other service the API and the worker need.
//
// Start calls Connect with the configuration, before it configures the plugins, only if the API or the worker server
// runs, then Close when the server stops. Connect does not return if the connection fails: it logs the error and exits or panics.
type Connector interface {
	Plugin
	Connect(configuration *config.Configuration)
	Close()
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

	server.API.Functions.NewMessageFromQueueCache = func(message string) bool {
		return true
	}
	server.API.Functions.BuildTokenFromID = func(ctx context.Context, tokenID string) (*authentication.Token, error) {
		return &authentication.Token{}, nil
	}
	server.API.Functions.ExtraPropertiesForTokenWithID = func(tokenID string) (map[string]string, error) {
		return map[string]string{}, nil
	}
	server.API.Functions.ExtendTokenExpirationDateFromID = func(ctx context.Context, tokenID string, ttl time.Duration) error {
		return nil
	}
	server.API.Functions.MiddlewareIsValidAuthToken = func(c *gin.Context, tokenID string) (bool, error) {
		return false, nil
	}
	server.API.Functions.MiddlewareGetUserAbilities = func(c *gin.Context, userID string) (bool, []string) {
		return false, []string{}
	}

	return server
}

// AddPlugin adds a plugin, which is configured when the server starts.
func (s *Server) AddPlugin(p Plugin) {
	s.plugins = append(s.plugins, p)
}
