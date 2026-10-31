// Package kafka provides a Kafka client to produce and consume messages.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/random"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// Client is a Kafka client, which may be a message producer. It is a broker.Broker.
type Client struct {
	// Host and Port are the address of a broker, used by Start.
	Host string
	Port int

	HasProducer  bool
	Producer     sarama.AsyncProducer
	TopicsPrefix string

	// Options are the settings of the connection. They are read by Initialize.
	Options Options

	// MaxAttempts is the number of times a message is handled before it is dead-lettered, 1 or more.
	MaxAttempts int
	// RetryDelay is the delay before the second attempt. It doubles at each attempt, up to a minute.
	RetryDelay time.Duration

	// Deduplicator remembers the messages handled, so that a message delivered twice, after a crash or a rebalance, is
	// not handled twice. Nothing is remembered if it is nil.
	Deduplicator broker.Deduplicator

	producerMu   sync.RWMutex
	producerOnce sync.Once

	// outbox holds the messages to send, in order, until the producer takes them. It is closed by Close.
	outbox       chan *sarama.ProducerMessage
	outboxMu     sync.RWMutex
	outboxClosed bool
	senderDone   chan struct{} // Closed when the messages of the outbox are given to the producer.

	stop       context.CancelFunc // Stops the consumers, see Close.
	background sync.WaitGroup     // The consumers and the connection of the producer.
	stats      stats
}

var _ broker.Broker = (*Client)(nil)

// Defaults of the Client.
const (
	DefaultMaxAttempts = 3
	DefaultRetryDelay  = time.Second

	// outboxSize is the number of messages waiting for the producer, which is connected in the background.
	outboxSize = 1024

	// maxRetryDelay is the longest delay between two attempts.
	maxRetryDelay = time.Minute

	// messageIDLength is the length of the identifier of a message.
	messageIDLength = 24

	// deadLetterSuffix is appended to the name of a topic to get the topic of its dead-lettered messages.
	deadLetterSuffix = "_dlq"
)

// DeadLetterTopic returns the name of the topic of the dead letters of a topic.
func DeadLetterTopic(topic string) string {
	return topic + deadLetterSuffix
}

// Replaced by the tests.
var (
	// newConsumerGroup creates a consumer group.
	newConsumerGroup = sarama.NewConsumerGroup

	// newAsyncProducer creates the producer.
	newAsyncProducer = sarama.NewAsyncProducer

	// retryDelay is the delay before a consumer tries again to connect or to consume.
	retryDelay = time.Second

	// closeTimeout is the longest wait for the consumers to stop when the client is closed.
	closeTimeout = 30 * time.Second
)

// NewClient returns a Client ready to be initialized.
func NewClient() *Client {
	return &Client{
		MaxAttempts: DefaultMaxAttempts,
		RetryDelay:  DefaultRetryDelay,
		outbox:      make(chan *sarama.ProducerMessage, outboxSize),
		senderDone:  make(chan struct{}),
	}
}

// Start connects the client to the broker at Host and Port, as a producer, and starts the consumers: see Initialize.
// It makes Client a broker.Broker.
func (client *Client) Start(consumers []broker.MessageConsumer) {
	client.Initialize(client.Host, client.Port, client.TopicsPrefix, true, consumers)
}

// Initialize connects the client to Kafka and starts the producer, if hasProducer
// is true, and one consumer per broker.MessageConsumer, each in its own goroutine and its own consumer group.
// Topics and groups are prefixed with prefix. Connections are retried every second until they succeed.
func (client *Client) Initialize(host string, port int, prefix string, hasProducer bool, consumers []broker.MessageConsumer) {
	client.TopicsPrefix = prefix
	client.HasProducer = hasProducer

	addresses := []string{fmt.Sprintf("%s:%d", host, port)}
	config := client.Options.NewConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	client.stop = stop

	if client.HasProducer {
		client.background.Add(1)
		go func() {
			defer client.background.Done()
			client.connectProducer(ctx, addresses, config)
		}()
	}

	// The consumers are stopped by Close or by an interruption.
	for _, consumer := range consumers {
		client.background.Add(1)
		go func() {
			defer client.background.Done()
			client.consume(ctx, addresses, config, consumer)
		}()
	}

	if client.Options.StatsInterval > 0 {
		go client.logStats(ctx, client.Options.StatsInterval)
	}
}

// connectProducer creates the producer, trying again every second, then attaches it to the client.
func (client *Client) connectProducer(ctx context.Context, addresses []string, config *sarama.Config) {
	for {
		producer, err := newAsyncProducer(addresses, config)
		if err == nil {
			log.WithFields(logrus.Fields{
				"addresses": addresses,
				"type":      "producer",
				"prefix":    client.TopicsPrefix,
			}).Info("connected to kafka")
			client.attachProducer(producer)

			return
		}

		log.WithFields(logrus.Fields{
			"addresses": addresses,
			"type":      "producer",
			"error":     err.Error(),
		}).Error("unable to connect to kafka")

		if !sleep(ctx, retryDelay) {
			return
		}
	}
}

// UseProducer makes the client a producer with the producer of the caller, such as a mock in the tests, instead of
// the one Initialize connects.
func (client *Client) UseProducer(producer sarama.AsyncProducer) {
	client.HasProducer = true
	client.attachProducer(producer)
}

// attachProducer starts giving the messages of the outbox to the producer, and tracks its acknowledgements.
func (client *Client) attachProducer(producer sarama.AsyncProducer) {
	client.producerMu.Lock()
	client.Producer = producer
	client.producerMu.Unlock()

	go func() {
		defer close(client.senderDone)

		for message := range client.outbox {
			producer.Input() <- message
		}
	}()

	go func() {
		for message := range producer.Successes() {
			client.stats.sent.Add(1)
			log.WithFields(logrus.Fields{
				"type":      "producer",
				"topic":     message.Topic,
				"partition": message.Partition,
				"offset":    message.Offset,
			}).Info("kafka message sent")
		}
	}()

	go func() {
		for err := range producer.Errors() {
			client.stats.sendErrors.Add(1)
			log.WithFields(logrus.Fields{
				"type":  "producer",
				"topic": err.Msg.Topic,
				"error": err.Err.Error(),
			}).Error("kafka send message error")
		}
	}()
}

// UseDeduplicator sets the Deduplicator which makes the consumers ignore the messages already handled. It must be
// called before Start. It makes the client a broker.DeduplicatorUser.
func (client *Client) UseDeduplicator(d broker.Deduplicator) {
	client.Deduplicator = d
}

// Stats returns the counters of the client.
func (client *Client) Stats() Stats {
	return client.stats.snapshot()
}

// logStats logs the statistics every interval, until ctx is done.
func (client *Client) logStats(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := client.Stats()
			fields := logrus.Fields{
				"received":      stats.Received,
				"handled":       stats.Handled,
				"retried":       stats.Retried,
				"duplicates":    stats.Duplicates,
				"dead_lettered": stats.DeadLettered,
				"sent":          stats.Sent,
				"send_errors":   stats.SendErrors,
				"dropped":       stats.Dropped,
				"waiting":       len(client.outbox),
			}
			for _, partition := range slices.Sorted(maps.Keys(stats.Lag)) {
				fields["lag "+partition] = stats.Lag[partition]
			}
			log.WithFields(fields).Info("kafka stats")

		case <-ctx.Done():
			return
		}
	}
}

// Close stops the consumers, waits for them to leave their group, then sends the messages still waiting and closes
// the producer. The message being handled is finished, and committed, first: the consumers have closeTimeout to
// stop, and so has the producer to send the messages.
func (client *Client) Close() {
	if client.stop != nil {
		client.stop()
	}

	if !waitFor(client.background.Wait) {
		log.Warn("the kafka consumers did not stop in time")
	}

	client.closeOutbox()
	if client.producer() == nil {
		if waiting := len(client.outbox); waiting > 0 {
			client.stats.dropped.Add(uint64(waiting))
			log.WithField("count", waiting).Error("kafka messages dropped: the producer was never connected")
		}

		return
	}

	if !waitFor(func() { <-client.senderDone }) {
		log.Warn("the kafka producer did not send the waiting messages in time")
	}
	client.closeProducer()
}

// waitFor calls wait, and reports whether it returned within closeTimeout.
func waitFor(wait func()) bool {
	done := make(chan struct{})
	go func() {
		wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(closeTimeout):
		return false
	}
}

// closeOutbox stops accepting messages to send.
func (client *Client) closeOutbox() {
	client.outboxMu.Lock()
	defer client.outboxMu.Unlock()

	if !client.outboxClosed {
		client.outboxClosed = true
		close(client.outbox)
	}
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

// SendMessageToTopic sends a message to the topic, asynchronously, without a key: see SendKeyedMessageToTopic.
func (client *Client) SendMessageToTopic(message, topic string) {
	client.SendKeyedMessageToTopic("", message, topic)
}

// SendKeyedMessageToTopic sends a message to the topic, asynchronously. The messages are sent in the order of the
// calls, as soon as the producer is connected. It logs an error and does nothing if the client has no producer, if it
// is closed or if too many messages wait for the producer.
//
// The messages with the same key go to the same partition, so a consumer handles them in the order they are sent:
// use the identifier of an entity, such as a user, as the key. A message without a key goes to any partition.
//
// Each message gets an identifier, in the header HeaderMessageID, which the consumers use to ignore a message
// delivered twice, see broker.Deduplicator.
func (client *Client) SendKeyedMessageToTopic(key, message, topic string) {
	if !client.HasProducer {
		log.WithFields(logrus.Fields{
			"error": "not a producer",
		}).Error("can not send message")
		return
	}

	fullTopicName := client.fullTopicName(topic)
	producerMessage := &sarama.ProducerMessage{
		Topic:   fullTopicName,
		Value:   sarama.StringEncoder(message),
		Headers: []sarama.RecordHeader{{Key: []byte(HeaderMessageID), Value: []byte(random.String(messageIDLength))}},
	}
	if key != "" {
		producerMessage.Key = sarama.StringEncoder(key)
	}

	client.outboxMu.RLock()
	defer client.outboxMu.RUnlock()

	if client.outboxClosed {
		client.stats.dropped.Add(1)
		log.WithField("topic", fullTopicName).Error("kafka message dropped: the client is closed")

		return
	}

	select {
	case client.outbox <- producerMessage:
		// The body is not logged above the debug level: it holds secrets, such as the keys of the confirmation links.
		log.WithFields(logrus.Fields{"topic": fullTopicName, "size": len(message)}).WithField("message", message).Debug("kafka message queued")
	default:
		client.stats.dropped.Add(1)
		log.WithField("topic", fullTopicName).Error("kafka message dropped: too many messages wait for the producer")
	}
}

// producer returns the producer, or nil if it is not connected yet.
func (client *Client) producer() sarama.AsyncProducer {
	client.producerMu.RLock()
	defer client.producerMu.RUnlock()

	return client.Producer
}

// consume joins the group of the consumer, then handles the messages of its topic until ctx is done. The connection to
// Kafka is retried every second.
func (client *Client) consume(ctx context.Context, addresses []string, config *sarama.Config, mc broker.MessageConsumer) {
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
	var dedup broker.Deduplicator
	if !mc.Idempotent {
		dedup = client.Deduplicator
	}
	handler := groupHandler{
		topic:    topic,
		group:    groupName,
		action:   mc.Action,
		attempts: client.MaxAttempts,
		delay:    client.RetryDelay,
		stats:    &client.stats,
		dedup:    dedup,
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
		client.stats.dropped.Add(1)
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
		Topic: DeadLetterTopic(message.Topic),
		Key:   sarama.ByteEncoder(message.Key),
		Value: sarama.ByteEncoder(message.Value),
	}
	for _, header := range message.Headers {
		if string(header.Key) == HeaderMessageID {
			headers[HeaderMessageID] = string(header.Value)
		}
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
