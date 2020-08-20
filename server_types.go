package ginamite

import (
	"time"

	"github.com/cygy/ginamite/api"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/web"
	"github.com/cygy/ginamite/worker"

	"github.com/gin-gonic/gin"
)

// Server : represents the main struct of the application containing all the settings.
type Server struct {
	ApplicationConfiguration ApplicationConfiguration
	Worker                   *worker.Server
	API                      *api.Server
	Web                      *web.Server

	// private properties
	plugins []Plugin
}

// SpecializedServer : abstract defintion of a API/Worker/Web specialized server.
type SpecializedServer interface {
	CreateKafkaConsumers() []kafka.MessageConsumer
	Start()
}

// Plugin : interface that loads some configuration ot a server.
type Plugin interface {
	Configure(s *Server)
}

// ApplicationConfiguration : additional and optional configuration of te application.
type ApplicationConfiguration struct {
	FileRequired   bool
	InitializeFunc func(filePath string)
}

// NewServer : returns a newly created 'Server' struct.
func NewServer() *Server {
	server := &Server{}
	server.ApplicationConfiguration.FileRequired = false
	server.plugins = []Plugin{}

	server.Web = web.NewServer()
	server.Worker = worker.NewServer()

	server.API = api.NewServer()
	server.API.Functions.NewMessageFromQueueCache = func(message string) bool {
		return true
	}
	server.API.Functions.BuildTokenFromID = func(tokenID string) (*authentication.Token, error) {
		return &authentication.Token{}, nil
	}
	server.API.Functions.ExtraPropertiesForTokenWithID = func(tokenID string) (map[string]string, error) {
		return map[string]string{}, nil
	}
	server.API.Functions.ExtendTokenExpirationDateFromID = func(tokenID string, ttl time.Duration) error {
		return nil
	}
	server.API.Functions.MiddlewareIsValidAuthToken = func(c *gin.Context, tokenID string) bool {
		return false
	}
	server.API.Functions.MiddlewareGetUserAbilities = func(c *gin.Context, userID string) (bool, []string) {
		return false, []string{}
	}

	return server
}
