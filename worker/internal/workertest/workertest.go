// Package workertest provides helpers shared by the tests of the worker: a logger collecting its entries and a
// capture of the messages sent to Kafka.
package workertest

import (
	"encoding/json"
	"testing"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

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

// Queue is the capture of the messages sent to the broker.
type Queue struct {
	// Sender is the queue sending its messages to the capture.
	Sender *queue.Queue

	t      *testing.T
	broker *broker.Memory
	next   int
}

// Message is a message sent to the broker.
type Message struct {
	Topic   string
	Type    string
	Payload map[string]any
}

// CaptureQueue returns a Queue capturing the messages sent by its Sender.
func CaptureQueue(t *testing.T) *Queue {
	t.Helper()

	memory := broker.NewMemory()

	return &Queue{Sender: queue.New(memory), t: t, broker: memory}
}

// Next returns the next message sent to the broker, and ok is false if there is none.
func (q *Queue) Next() (message Message, ok bool) {
	q.t.Helper()

	sent := q.broker.Sent()
	if q.next >= len(sent) {
		return message, false
	}

	var m queue.Message
	if err := json.Unmarshal([]byte(sent[q.next].Message), &m); err != nil {
		q.t.Fatal(err)
	}

	message = Message{Topic: sent[q.next].Topic, Type: m.Type}
	message.Payload, _ = m.Payload.(map[string]any)
	q.next++

	return message, true
}

// Empty fails the test if a message has been sent to Kafka.
func (q *Queue) Empty() {
	q.t.Helper()

	if m, ok := q.Next(); ok {
		q.t.Errorf("unexpected message %+v", m)
	}
}
