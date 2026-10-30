// Package ginamite is a framework to build an application made of an API server, a worker server and a web server.
//
// A Server holds the three specialized servers and the settings of the application. The command line arguments choose
// the one which runs, see the package flags.
//
// The storage is provided by plugins, see Plugin and Connector: the core does not depend on a database.
package ginamite

import (
	"context"
	"time"

	"github.com/cygy/ginamite/api"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/web"
	"github.com/cygy/ginamite/worker"

	"github.com/gin-gonic/gin"
)

// Server is the main struct of the application, containing all the settings.
type Server struct {
	ApplicationConfiguration ApplicationConfiguration
	Worker                   *worker.Server
	API                      *api.Server
	Web                      *web.Server

	// private properties
	plugins []Plugin
}

// SpecializedServer is the definition of a specialized server: API, worker or web.
type SpecializedServer interface {
	CreateKafkaConsumers() []kafka.MessageConsumer
	Start()
}

// Plugin is a plugin which configures a server, such as the functions of the API and of the workers.
type Plugin interface {
	Configure(s *Server)
}

// Connector is a Plugin which holds a connection to a database, or to any other service the API and the worker need.
//
// Start calls Connect, before it configures the plugins, only if the API or the worker server runs, then Close when
// the server stops. Connect does not return if the connection fails: it logs the error and exits or panics.
type Connector interface {
	Plugin
	Connect()
	Close()
}

// NewServer returns a Server whose API functions are placeholders which accept nothing: a plugin or the application
// is expected to replace them.
func NewServer() *Server {
	server := &Server{
		plugins: []Plugin{},
		Web:     web.NewServer(),
		Worker:  worker.NewServer(),
		API:     api.NewServer(),
	}
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
