// Package kafka provides a Kafka client to produce and consume messages.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/random"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// Client is a Kafka client, which may be a message producer.
type Client struct {
	HasProducer     bool
	Producer        sarama.AsyncProducer
	ProducerSignals chan os.Signal
	TopicsPrefix    string

	// MaxAttempts is the number of times a message is handled before it is dead-lettered, 1 or more.
	MaxAttempts int
	// RetryDelay is the delay before the second attempt. It doubles at each attempt, up to a minute.
	RetryDelay time.Duration

	producerMu   sync.RWMutex
	producerOnce sync.Once
	stop         context.CancelFunc // Stops the consumers, see Close.
	consumers    sync.WaitGroup
}

// Defaults of the Client.
const (
	DefaultMaxAttempts = 3
	DefaultRetryDelay  = time.Second

	// maxRetryDelay is the longest delay between two attempts.
	maxRetryDelay = time.Minute

	// deadLetterSuffix is appended to the name of a topic to get the topic of its dead-lettered messages.
	deadLetterSuffix = "_dlq"
)

// MessageConsumer consumes the messages of a topic and executes Action on each one.
//
// The consumers of the same Group share the messages of the topic: each message is handled by one of them, so the
// workers sharing a group share the load. The consumers of different groups all receive every message, see UniqueGroup.
type MessageConsumer struct {
	Topic string
	// Group is the name of the consumer group, prefixed like the topics. It is the name of the topic if empty.
	Group string
	// Action handles a message and returns an error if it failed. The message is committed once Action returns: it is
	// handled again by the group if the process stops before, so Action must tolerate a message handled twice.
	//
	// A failed message is handled again, up to Client.MaxAttempts times, then it is sent to the topic of the dead
	// letters, the name of the topic with the suffix "_dlq", and the next message is handled. An error marked with
	// Permanent is not retried. A panic is recovered and handled like an error.
	Action func(value []byte) error
}

// Replaced by the tests.
var (
	// newConsumerGroup creates a consumer group.
	newConsumerGroup = sarama.NewConsumerGroup

	// retryDelay is the delay before a consumer tries again to connect or to consume.
	retryDelay = time.Second

	// closeTimeout is the longest wait for the consumers to stop when the client is closed.
	closeTimeout = 30 * time.Second
)

// UniqueGroup returns the name of a group no other process uses, so its consumer receives every message of the topic:
// use it for the messages every instance of a server must handle, such as the invalidations of a cache.
func UniqueGroup(name string) string {
	return name + "_" + random.String(8)
}

// NewClient returns a Client ready to be initialized.
func NewClient() *Client {
	return &Client{ProducerSignals: make(chan os.Signal, 1), MaxAttempts: DefaultMaxAttempts, RetryDelay: DefaultRetryDelay}
}

// Initialize connects the client to Kafka and starts the producer, if hasProducer
// is true, and one consumer per MessageConsumer, each in its own goroutine and its own consumer group.
// Topics and groups are prefixed with prefix. Connections are retried every second until they succeed.
func (client *Client) Initialize(host string, port int, prefix string, hasProducer bool, consumers []MessageConsumer) {
	// Set the shared vars.
	client.TopicsPrefix = prefix
	client.HasProducer = hasProducer

	addresses := []string{fmt.Sprintf("%s:%d", host, port)}

	// Create the kafka configuration.
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // A new group handles the messages sent from now on.

	// Loop for the producer.
	createProducer := func() {
		var wg sync.WaitGroup

		// Create the producer.
		for client.Producer == nil {
			if p, err := sarama.NewAsyncProducer(addresses, config); err != nil {
				log.WithFields(logrus.Fields{
					"host":  host,
					"port":  port,
					"type":  "producer",
					"error": err.Error(),
				}).Error("unable to connect to kafka")
				time.Sleep(1 * time.Second)
			} else {
				client.producerMu.Lock()
				client.Producer = p
				client.producerMu.Unlock()
				log.WithFields(logrus.Fields{
					"host":  host,
					"port":  port,
					"type":  "producer",
					"topic": client.TopicsPrefix,
				}).Info("connected to kafka")
			}
		}

		// Trap SIGINT to trigger a graceful shutdown.
		signal.Notify(client.ProducerSignals, os.Interrupt)

		// Track the successes and the errors.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for message := range client.producer().Successes() {
				log.WithFields(logrus.Fields{
					"type":      "producer",
					"value":     message.Value,
					"topic":     message.Topic,
					"partition": message.Partition,
				}).Info("kafka send message success")
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for err := range client.producer().Errors() {
				log.WithFields(logrus.Fields{
					"type":  "producer",
					"error": err.Error(),
				}).Error("kafka send message error")
			}
		}()

		// Wait the shutdown.
		wg.Wait()
	}

	// Create the producer.
	if client.HasProducer {
		go createProducer()
	}

	// Create the consumers, stopped by Close or by an interruption.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	client.stop = stop
	for _, consumer := range consumers {
		client.consumers.Add(1)
		go func() {
			defer client.consumers.Done()
			client.consume(ctx, addresses, config, consumer)
		}()
	}
}

// Close stops the consumers, waits for them to leave their group, then flushes and closes the producer. The message
// being handled is finished, and committed, first: the consumers have closeTimeout to stop.
func (client *Client) Close() {
	if client.stop != nil {
		client.stop()
	}

	stopped := make(chan struct{})
	go func() {
		client.consumers.Wait()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(closeTimeout):
		log.Warn("the kafka consumers did not stop in time")
	}

	client.closeProducer()
}

// closeProducer flushes and closes the producer, if it is connected, once.
func (client *Client) closeProducer() {
	client.producerOnce.Do(func() {
		producer := client.producer()
		if producer == nil {
			return
		}

		if err := producer.Close(); err != nil {
			log.WithField("error", err).Error("unable to close the kafka producer")
		}
	})
}

// consume joins the group of the consumer, then handles the messages of its topic until ctx is done. The connection to
// Kafka is retried every second.
func (client *Client) consume(ctx context.Context, addresses []string, config *sarama.Config, mc MessageConsumer) {
	topic := client.fullTopicName(mc.Topic)
	groupName := mc.Topic
	if mc.Group != "" {
		groupName = mc.Group
	}
	groupName = client.fullTopicName(groupName)

	var group sarama.ConsumerGroup
	for group == nil {
		g, err := newConsumerGroup(addresses, groupName, config)
		if err == nil {
			group = g
			break
		}

		log.WithFields(logrus.Fields{
			"addresses": addresses,
			"type":      "consumer",
			"group":     groupName,
			"error":     err.Error(),
		}).Error("unable to connect to kafka")

		if !sleep(ctx, retryDelay) {
			return
		}
	}

	defer func() {
		if err := group.Close(); err != nil {
			log.WithField("error", err).Error("unable to close the kafka consumer")
		}
	}()

	log.WithFields(logrus.Fields{
		"type":  "consumer",
		"topic": topic,
		"group": groupName,
	}).Info("connected to kafka")

	go func() {
		for err := range group.Errors() {
			log.WithFields(logrus.Fields{
				"type":  "consumer",
				"error": err.Error(),
			}).Error("kafka consumer error")
		}
	}()

	// Consume returns when the group is rebalanced: it must be called again.
	handler := groupHandler{
		topic:    topic,
		group:    groupName,
		action:   mc.Action,
		attempts: client.MaxAttempts,
		delay:    client.RetryDelay,
		deadLetter: func(message *sarama.ConsumerMessage, err error, attempts int) {
			client.sendDeadLetter(ctx, message, groupName, err, attempts)
		},
	}
	for ctx.Err() == nil {
		err := group.Consume(ctx, []string{topic}, handler)
		if err == nil || ctx.Err() != nil {
			continue
		}
		if errors.Is(err, sarama.ErrClosedConsumerGroup) {
			return
		}

		log.WithFields(logrus.Fields{
			"type":  "consumer",
			"topic": topic,
			"error": err.Error(),
		}).Error("kafka consume error")

		if !sleep(ctx, retryDelay) {
			return
		}
	}
}

// sleep waits for d, and reports false if ctx is done first.
func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}

// SendMessageToTopic sends a message to the topic, asynchronously.
// It logs an error and does nothing if the client has no producer.
func (client *Client) SendMessageToTopic(message, topic string) {
	if !client.HasProducer {
		log.WithFields(logrus.Fields{
			"error": "not a producer",
		}).Error("can not send message")
		return
	}

	go func() {
		producer := client.producer()
		if producer == nil {
			log.WithField("error", "producer not connected").Error("can not send message")
			return
		}

		fullTopicName := client.fullTopicName(topic)
		producerMessage := &sarama.ProducerMessage{
			Topic: fullTopicName,
			Value: sarama.StringEncoder(message),
		}

		select {
		case producer.Input() <- producerMessage:
			// The body is not logged above the debug level: it holds secrets, such as the keys of the confirmation links.
			fields := logrus.Fields{
				"topic": fullTopicName,
				"size":  len(message),
			}
			log.WithFields(fields).Info("kafka message sent")
			log.WithFields(fields).WithField("message", message).Debug("kafka message body")

		case <-client.ProducerSignals:
			client.closeProducer()
		}
	}()
}

// producer returns the producer, or nil if it is not connected yet.
func (client *Client) producer() sarama.AsyncProducer {
	client.producerMu.RLock()
	defer client.producerMu.RUnlock()

	return client.Producer
}

// sendDeadLetter sends a message no attempt could handle to the topic of the dead letters, with the reason in its
// headers. The message is lost, and logged, if the client has no connected producer.
func (client *Client) sendDeadLetter(ctx context.Context, message *sarama.ConsumerMessage, group string, cause error, attempts int) {
	fields := logrus.Fields{
		"topic":     message.Topic,
		"group":     group,
		"partition": message.Partition,
		"offset":    message.Offset,
		"attempts":  attempts,
		"error":     cause.Error(),
	}

	producer := client.producer()
	if producer == nil {
		log.WithFields(fields).Error("kafka message dropped: no producer for the dead letters")
		return
	}

	headers := map[string]string{
		"error":     cause.Error(),
		"attempts":  strconv.Itoa(attempts),
		"group":     group,
		"topic":     message.Topic,
		"partition": strconv.Itoa(int(message.Partition)),
		"offset":    strconv.FormatInt(message.Offset, 10),
	}
	deadLetter := &sarama.ProducerMessage{
		Topic: message.Topic + deadLetterSuffix,
		Key:   sarama.ByteEncoder(message.Key),
		Value: sarama.ByteEncoder(message.Value),
	}
	for name, value := range headers {
		deadLetter.Headers = append(deadLetter.Headers, sarama.RecordHeader{Key: []byte(name), Value: []byte(value)})
	}

	select {
	case producer.Input() <- deadLetter:
		log.WithFields(fields).Error("kafka message dead-lettered")
	case <-ctx.Done():
		// The message is not committed: it is handled again.
	}
}

// fullTopicName returns the name of the topic prefixed with the topics prefix.
func (client *Client) fullTopicName(topic string) string {
	return client.TopicsPrefix + topic
}
