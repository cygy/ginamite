package kafka

import (
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
)

// Service is the name of Kafka as a broker.Service of the configuration.
const Service = "kafka"

// Plugin makes Kafka the message broker of the servers: it creates the topics if the configuration asks for it, and
// sets a Client as the broker, see ginamite.Server.UseBroker, which the server starts with the consumers of the chosen
// server. The configuration chooses it with `broker.service: kafka` and gives the address of the broker, otherwise the plugin
// does nothing.
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
	if !enabled(s.Config) {
		return
	}

	options, err := optionsFrom(s.Config)
	if err != nil {
		log.WithField("error", err).Fatal("invalid kafka configuration")
		return
	}

	createTopics(s.Config, s.Config.Broker.Host, s.Config.Broker.Port, options)

	client := NewClient()
	client.Host, client.Port = s.Config.Broker.Host, s.Config.Broker.Port
	client.TopicsPrefix = s.Config.Kafka.TopicsPrefix
	client.Options = options
	client.MaxAttempts = max(s.Config.Kafka.MaxAttempts, 1)
	client.RetryDelay = time.Duration(s.Config.Kafka.RetryDelay) * time.Millisecond
	s.UseBroker(client)
}

// enabled reports whether the configuration chooses Kafka as the broker, and has its address.
func enabled(c *config.Configuration) bool {
	return c.Broker.Service == Service && c.Broker.Host != "" && c.Broker.Port > 0
}
