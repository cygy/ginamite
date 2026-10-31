package kafka

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cygy/ginamite/common/log"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// TopicSpec describes the topics to create.
type TopicSpec struct {
	Partitions        int32
	ReplicationFactor int16

	// Retention is how long the messages are kept, and DeadLetterRetention how long the dead letters are. Kafka keeps
	// them as long as its own setting if 0.
	Retention           time.Duration
	DeadLetterRetention time.Duration
}

// topicCreator creates topics.
type topicCreator interface {
	CreateTopic(topic string, detail *sarama.TopicDetail, validateOnly bool) error
	Close() error
}

// newTopicCreator connects to the brokers; the tests replace it.
var newTopicCreator = func(addresses []string, config *sarama.Config) (topicCreator, error) {
	return sarama.NewClusterAdmin(addresses, config)
}

// EnsureTopics creates, with the settings of the spec, the topics, and the topics of their dead letters, which do not
// exist yet; the prefix is prepended to their names. The topics which exist are not changed. It tries to create all
// the topics, and returns all the errors.
func EnsureTopics(host string, port int, options Options, prefix string, topics []string, spec TopicSpec) error {
	admin, err := newTopicCreator([]string{fmt.Sprintf("%s:%d", host, port)}, options.NewConfig())
	if err != nil {
		return fmt.Errorf("unable to connect to kafka: %w", err)
	}
	defer func() { _ = admin.Close() }()

	var errs []error
	for _, topic := range topics {
		for _, deadLetters := range []bool{false, true} {
			name, retention := prefix+topic, spec.Retention
			if deadLetters {
				name, retention = DeadLetterTopic(name), spec.DeadLetterRetention
			}

			if err := createTopic(admin, name, spec, retention); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// createTopic creates a topic, if it does not exist.
func createTopic(admin topicCreator, name string, spec TopicSpec, retention time.Duration) error {
	detail := &sarama.TopicDetail{
		NumPartitions:     max(spec.Partitions, 1),
		ReplicationFactor: max(spec.ReplicationFactor, 1),
	}
	if retention > 0 {
		detail.ConfigEntries = map[string]*string{"retention.ms": new(strconv.FormatInt(retention.Milliseconds(), 10))}
	}

	err := admin.CreateTopic(name, detail, false)
	switch {
	case err == nil:
		log.WithFields(logrus.Fields{"topic": name, "partitions": detail.NumPartitions}).Info("kafka topic created")
	case errors.Is(err, sarama.ErrTopicAlreadyExists):
	default:
		return fmt.Errorf("unable to create the topic %s: %w", name, err)
	}

	return nil
}
