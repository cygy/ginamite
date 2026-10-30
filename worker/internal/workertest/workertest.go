// Package workertest provides helpers shared by the tests of the worker: a logger collecting its entries and a
// capture of the messages sent to Kafka.
package workertest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

	"github.com/IBM/sarama/mocks"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// Hook collects the entries of the main logger, which discards them.
var Hook *logtest.Hook

func init() {
	var logger *logrus.Logger
	logger, Hook = logtest.NewNullLogger()
	log.Main = logger.WithField("server", "test")
}

// maxMessages is the number of messages a Queue can capture.
const maxMessages = 10

// Queue is the capture of the messages sent to Kafka.
type Queue struct {
	t        *testing.T
	producer *mocks.AsyncProducer
}

// Message is a message sent to Kafka.
type Message struct {
	Topic   string
	Type    string
	Payload map[string]any
}

// CaptureQueue replaces the main Kafka client, for the duration of the test, with a client sending its messages to
// the returned Queue.
func CaptureQueue(t *testing.T) *Queue {
	t.Helper()

	config := mocks.NewTestConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewAsyncProducer(t, config)
	for range maxMessages {
		producer.ExpectInputAndSucceed()
	}

	client := kafka.NewClient()
	client.HasProducer = true
	client.Producer = producer

	old := kafka.Main
	kafka.Main = client
	t.Cleanup(func() { kafka.Main = old })

	return &Queue{t: t, producer: producer}
}

// Next returns the next message sent to Kafka, and ok is false if there is none after a short delay.
func (q *Queue) Next() (message Message, ok bool) {
	q.t.Helper()

	select {
	case sent := <-q.producer.Successes():
		value, err := sent.Value.Encode()
		if err != nil {
			q.t.Fatal(err)
		}

		var m queue.Message
		if err := json.Unmarshal(value, &m); err != nil {
			q.t.Fatal(err)
		}

		message = Message{Topic: sent.Topic, Type: m.Type}
		message.Payload, _ = m.Payload.(map[string]any)

		return message, true
	case <-time.After(200 * time.Millisecond):
		return message, false
	}
}

// Empty fails the test if a message has been sent to Kafka.
func (q *Queue) Empty() {
	q.t.Helper()

	if m, ok := q.Next(); ok {
		q.t.Errorf("unexpected message %+v", m)
	}
}
