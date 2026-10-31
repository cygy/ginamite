package kafka

import (
	"crypto/tls"
	"time"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

// SASL mechanisms.
const (
	SASLPlain       = sarama.SASLTypePlaintext
	SASLSCRAMSHA256 = sarama.SASLTypeSCRAMSHA256
	SASLSCRAMSHA512 = sarama.SASLTypeSCRAMSHA512
)

// Options are the settings of the connection to Kafka.
type Options struct {
	// Version is the version of the brokers. It is the one of the library if zero.
	Version sarama.KafkaVersion

	// TLS encrypts the connection if not nil.
	TLS *tls.Config

	// SASL authenticates the client if its mechanism is not empty.
	SASL SASL

	// StatsInterval is the interval of the log of the statistics, see Client.Stats. They are not logged if 0.
	StatsInterval time.Duration
}

// SASL holds the credentials of the client.
type SASL struct {
	// Mechanism is SASLPlain, SASLSCRAMSHA256 or SASLSCRAMSHA512.
	Mechanism string
	Username  string
	Password  string
}

// NewConfig returns the configuration of the library: a producer which waits for all the replicas and does not
// duplicate a message when it tries again, and consumers starting from the newest message when their group is new.
func (o Options) NewConfig() *sarama.Config {
	config := sarama.NewConfig()

	if o.Version != (sarama.KafkaVersion{}) {
		config.Version = o.Version
	}

	// A message is acknowledged by all the in-sync replicas, and sent again if it fails, without duplicate nor
	// reordering: this is what the idempotent producer of the library requires.
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Idempotent = true
	config.Producer.Retry.Max = 5
	config.Net.MaxOpenRequests = 1
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // A new group handles the messages sent from now on.

	if o.TLS != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = o.TLS
	}

	if o.SASL.Mechanism != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLMechanism(o.SASL.Mechanism)
		config.Net.SASL.User = o.SASL.Username
		config.Net.SASL.Password = o.SASL.Password

		switch o.SASL.Mechanism {
		case SASLSCRAMSHA256:
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &scramClient{hash: scram.SHA256} }
		case SASLSCRAMSHA512:
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &scramClient{hash: scram.SHA512} }
		}
	}

	return config
}

// scramClient is the SCRAM client of the SASL authentication.
type scramClient struct {
	hash         scram.HashGeneratorFcn
	conversation *scram.ClientConversation
}

func (c *scramClient) Begin(userName, password, authzID string) error {
	client, err := c.hash.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}

	c.conversation = client.NewConversation()

	return nil
}

func (c *scramClient) Step(challenge string) (string, error) { return c.conversation.Step(challenge) }

func (c *scramClient) Done() bool { return c.conversation.Done() }
