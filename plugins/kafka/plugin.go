package kafka

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
)

// Service is the name of Kafka as a broker.Service of the configuration.
const Service = "kafka"

// Section is the name of the section of the configuration holding the settings, see Settings.
const Section = "kafka"

// Plugin makes Kafka the message broker of the servers: it creates the topics if the configuration asks for it, and
// sets a Client as the broker, see ginamite.Server.UseBroker, which the server starts with the consumers of the chosen
// server. The configuration chooses it with `broker.service: kafka` and gives the address of the broker, otherwise the plugin
// does nothing.
//
// It also runs the command on the dead letters, see Command.
type Plugin struct {
	dlq dlqArguments
}

// NewPlugin returns a Plugin.
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Configure sets the Kafka client, built from the configuration, as the broker of the server. It returns an error if
// the Kafka settings are invalid.
func (p *Plugin) Configure(s *ginamite.Server) error {
	if !enabled(s.Config) {
		return nil
	}

	settings, err := settingsFrom(s.Config)
	if err != nil {
		return fmt.Errorf("invalid kafka configuration: %w", err)
	}

	options, err := optionsFrom(settings)
	if err != nil {
		return fmt.Errorf("invalid kafka configuration: %w", err)
	}

	createTopics(settings, s.Config.Broker.Host, s.Config.Broker.Port, options)

	client := NewClient()
	client.Host, client.Port = s.Config.Broker.Host, s.Config.Broker.Port
	client.TopicsPrefix = settings.TopicsPrefix
	client.Options = options
	client.MaxAttempts = max(settings.MaxAttempts, 1)
	client.RetryDelay = time.Duration(settings.RetryDelay) * time.Millisecond
	s.UseBroker(client)

	return nil
}

// enabled reports whether the configuration chooses Kafka as the broker, and has its address.
func enabled(c *config.Configuration) bool {
	return c.Broker.Service == Service && c.Broker.Host != "" && c.Broker.Port > 0
}
