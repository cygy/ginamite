package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

	"github.com/IBM/sarama"
)

// optionsFrom returns the options of the connection to Kafka from its settings.
func optionsFrom(k Settings) (Options, error) {
	var options Options

	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return options, fmt.Errorf("invalid kafka version %q: %w", k.Version, err)
		}

		options.Version = version
	}

	tlsConfig, err := k.TLS.TLSConfig()
	if err != nil {
		return options, err
	}
	options.TLS = tlsConfig

	if mechanism := strings.ToUpper(k.SASL.Mechanism); mechanism != "" {
		switch mechanism {
		case SASLPlain, SASLSCRAMSHA256, SASLSCRAMSHA512:
		default:
			return options, fmt.Errorf("invalid kafka SASL mechanism %q: use %s, %s or %s", k.SASL.Mechanism, SASLPlain, SASLSCRAMSHA256, SASLSCRAMSHA512)
		}
		if k.SASL.Username == "" {
			return options, fmt.Errorf("the kafka SASL mechanism %s needs a username", mechanism)
		}

		options.SASL = SASL{Mechanism: mechanism, Username: k.SASL.Username, Password: k.SASL.Password}
	}

	options.StatsInterval = time.Duration(k.StatsInterval) * time.Second

	return options, nil
}

// createTopics creates the topics which do not exist yet, if the settings ask for it. The failure is
// logged: the topics may be created by the brokers, or by the operators.
func createTopics(k Settings, host string, port int, options Options) {
	if !k.Topics.Create {
		return
	}

	spec := TopicSpec{
		Partitions:          k.Topics.Partitions,
		ReplicationFactor:   k.Topics.ReplicationFactor,
		Retention:           time.Duration(k.Topics.RetentionHours) * time.Hour,
		DeadLetterRetention: time.Duration(k.Topics.DeadLetterRetentionHours) * time.Hour,
	}
	if err := EnsureTopics(host, port, options, k.TopicsPrefix, queue.Topics(), spec); err != nil {
		log.Error("unable to create the kafka topics", "error", err.Error())
	}
}
