package kafka

import (
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/log"
)

// Plugin makes Kafka the message broker of the servers: it creates the topics if the configuration asks for it, and
// sets a Client as the broker, see ginamite.Server.UseBroker, which the server starts with the consumers of the chosen
// server. Without the service "kafka" in the configuration, there is no broker.
//
// It also runs the command on the dead letters, see Command.
type Plugin struct{}

// NewPlugin returns a Plugin.
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Configure sets the Kafka client, built from the configuration, as the broker of the server. It stops the process if
// the Kafka settings are invalid.
func (p *Plugin) Configure(s *ginamite.Server) {
	service, ok := s.Config.KafkaConfiguration()
	if !ok {
		return
	}

	options, err := optionsFrom(s.Config)
	if err != nil {
		log.WithField("error", err).Fatal("invalid kafka configuration")
		return
	}

	createTopics(s.Config, service.Host, service.Port, options)

	client := NewClient()
	client.Host, client.Port = service.Host, service.Port
	client.TopicsPrefix = s.Config.Kafka.TopicsPrefix
	client.Options = options
	client.MaxAttempts = max(s.Config.Kafka.MaxAttempts, 1)
	client.RetryDelay = time.Duration(s.Config.Kafka.RetryDelay) * time.Millisecond
	s.UseBroker(client)
}
