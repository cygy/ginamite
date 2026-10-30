// Package kafka provides a Kafka client to produce and consume messages.
package kafka

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/cygy/ginamite/common/log"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// Client is a Kafka client, which may be a message producer.
type Client struct {
	HasProducer     bool
	Producer        sarama.AsyncProducer
	ProducerSignals chan os.Signal
	TopicsPrefix    string
}

// MessageConsumer consumes the messages of a topic and executes Action on each one.
type MessageConsumer struct {
	Topic  string
	Action func(value []byte)
}

// NewClient returns a Client ready to be initialized.
func NewClient() *Client {
	return &Client{ProducerSignals: make(chan os.Signal, 1)}
}

// Initialize connects the client to Kafka and starts the producer, if hasProducer
// is true, and one consumer per MessageConsumer, each in its own goroutine.
// Topics are prefixed with prefix. Connections are retried every second until they succeed.
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

	// Loop for the consumers.
	createConsumer := func(mc MessageConsumer) {
		var (
			consumer          sarama.Consumer
			partitionConsumer sarama.PartitionConsumer
			consumerSignals   = make(chan os.Signal, 1)
			wg                sync.WaitGroup
		)

		fullTopicName := client.fullTopicName(mc.Topic)

		// Create the consumer.
		for consumer == nil {
			if c, err := sarama.NewConsumer(addresses, config); err != nil {
				log.WithFields(logrus.Fields{
					"host":  host,
					"port":  port,
					"type":  "consumer",
					"error": err.Error(),
				}).Error("unable to connect to kafka")
				time.Sleep(1 * time.Second)
			} else {
				log.WithFields(logrus.Fields{
					"host":  host,
					"port":  port,
					"type":  "consumer",
					"topic": fullTopicName,
				}).Info("connected to kafka")

				for partitionConsumer == nil {
					if pc, err := c.ConsumePartition(fullTopicName, 0, sarama.OffsetNewest); err != nil {
						log.WithFields(logrus.Fields{
							"type":  "consumer",
							"error": err.Error(),
						}).Error("kafka consume partition error")
						time.Sleep(1 * time.Second)
					} else {
						consumer = c
						partitionConsumer = pc
					}
				}
			}
		}

		// Trap SIGINT to trigger a graceful shutdown.
		signal.Notify(consumerSignals, os.Interrupt)

		// Track the messages and the errors.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for message := range partitionConsumer.Messages() {
				log.WithFields(logrus.Fields{
					"topic":   fullTopicName,
					"message": string(message.Value),
				}).Info("kafka message received")
				mc.Action(message.Value)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for err := range partitionConsumer.Errors() {
				log.WithFields(logrus.Fields{
					"type":  "consumer",
					"error": err.Error(),
				}).Error("kafka partition error")
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-consumerSignals
			partitionConsumer.AsyncClose()
			if err := consumer.Close(); err != nil {
				log.WithField("error", err).Error("unable to close the kafka consumer")
			}
		}()

		// Wait the shutdown.
		wg.Wait()
	}

	// Create the producer.
	if client.HasProducer {
		go createProducer()
	}

	// Create the consumers.
	for _, consumer := range consumers {
		go createConsumer(consumer)
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
