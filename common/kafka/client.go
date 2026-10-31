// Package kafka provides a Kafka client to produce and consume messages.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
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

	stop      context.CancelFunc // Stops the consumers, see Close.
	consumers sync.WaitGroup
}

// MessageConsumer consumes the messages of a topic and executes Action on each one.
//
// The consumers of the same Group share the messages of the topic: each message is handled by one of them, so the
// workers sharing a group share the load. The consumers of different groups all receive every message, see UniqueGroup.
type MessageConsumer struct {
	Topic string
	// Group is the name of the consumer group, prefixed like the topics. It is the name of the topic if empty.
	Group string
	// Action handles a message. The message is committed once Action returns: it is handled again by the group if the
	// process stops before, so Action must tolerate a message handled twice.
	Action func(value []byte)
}

// Replaced by the tests.
var (
	// newConsumerGroup creates a consumer group.
	newConsumerGroup = sarama.NewConsumerGroup

	// retryDelay is the delay before a consumer tries again to connect or to consume.
	retryDelay = time.Second
)

// UniqueGroup returns the name of a group no other process uses, so its consumer receives every message of the topic:
// use it for the messages every instance of a server must handle, such as the invalidations of a cache.
func UniqueGroup(name string) string {
	return name + "_" + random.String(8)
}

// NewClient returns a Client ready to be initialized.
func NewClient() *Client {
	return &Client{ProducerSignals: make(chan os.Signal, 1)}
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
				client.Producer = p
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
			for message := range client.Producer.Successes() {
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
			for err := range client.Producer.Errors() {
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

// Close stops the consumers and waits for them to leave their group. The message being handled is finished, and
// committed, first.
func (client *Client) Close() {
	if client.stop != nil {
		client.stop()
	}

	client.consumers.Wait()
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
	handler := groupHandler{topic: topic, action: mc.Action}
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
		fullTopicName := client.fullTopicName(topic)
		producerMessage := &sarama.ProducerMessage{
			Topic: fullTopicName,
			Value: sarama.StringEncoder(message),
		}

		select {
		case client.Producer.Input() <- producerMessage:
			log.WithFields(logrus.Fields{
				"topic":   fullTopicName,
				"message": message,
			}).Info("kafka message sent")

		case <-client.ProducerSignals:
			client.Producer.AsyncClose()
		}
	}()
}

// fullTopicName returns the name of the topic prefixed with the topics prefix.
func (client *Client) fullTopicName(topic string) string {
	return client.TopicsPrefix + topic
}
