package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// optionsFrom returns the options of the connection to Kafka from the configuration.
func optionsFrom(c *config.Configuration) (Options, error) {
	var options Options

	if c.Kafka.Version != "" {
		version, err := sarama.ParseKafkaVersion(c.Kafka.Version)
		if err != nil {
			return options, fmt.Errorf("invalid kafka version %q: %w", c.Kafka.Version, err)
		}

		options.Version = version
	}

	tlsConfig, err := c.Kafka.TLS.TLSConfig()
	if err != nil {
		return options, err
	}
	options.TLS = tlsConfig

	if mechanism := strings.ToUpper(c.Kafka.SASL.Mechanism); mechanism != "" {
		switch mechanism {
		case SASLPlain, SASLSCRAMSHA256, SASLSCRAMSHA512:
		default:
			return options, fmt.Errorf("invalid kafka SASL mechanism %q: use %s, %s or %s", c.Kafka.SASL.Mechanism, SASLPlain, SASLSCRAMSHA256, SASLSCRAMSHA512)
		}
		if c.Kafka.SASL.Username == "" {
			return options, fmt.Errorf("the kafka SASL mechanism %s needs a username", mechanism)
		}

		options.SASL = SASL{Mechanism: mechanism, Username: c.Kafka.SASL.Username, Password: c.Kafka.SASL.Password}
	}

	options.StatsInterval = time.Duration(c.Kafka.StatsInterval) * time.Second

	return options, nil
}

// createTopics creates the topics which do not exist yet, if the configuration asks for it. The failure is
// logged: the topics may be created by the brokers, or by the operators.
func createTopics(c *config.Configuration, host string, port int, options Options) {
	if !c.Kafka.Topics.Create {
		return
	}

	spec := TopicSpec{
		Partitions:          c.Kafka.Topics.Partitions,
		ReplicationFactor:   c.Kafka.Topics.ReplicationFactor,
		Retention:           time.Duration(c.Kafka.Topics.RetentionHours) * time.Hour,
		DeadLetterRetention: time.Duration(c.Kafka.Topics.DeadLetterRetentionHours) * time.Hour,
	}
	if err := EnsureTopics(host, port, options, c.Kafka.TopicsPrefix, queue.Topics(), spec); err != nil {
		log.WithFields(logrus.Fields{"error": err.Error()}).Error("unable to create the kafka topics")
	}
}
