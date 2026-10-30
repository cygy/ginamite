package kafka

import (
	"os"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

func TestNewClient(t *testing.T) {
	c := NewClient()

	if c.ProducerSignals == nil || cap(c.ProducerSignals) != 1 || c.HasProducer || c.Producer != nil {
		t.Errorf("NewClient() = %+v", c)
	}
}

func TestClient_fullTopicName(t *testing.T) {
	tests := []struct {
		prefix, topic, want string
	}{
		{"", "users", "users"},
		{"app_", "users", "app_users"},
	}

	for _, tt := range tests {
		c := &Client{TopicsPrefix: tt.prefix}
		if got := c.fullTopicName(tt.topic); got != tt.want {
			t.Errorf("fullTopicName(%q) = %q, want %q", tt.topic, got, tt.want)
		}
	}
}

func TestClient_SendMessageToTopic_NoProducer(t *testing.T) {
	hook := useHook(t)

	NewClient().SendMessageToTopic("hello", "topic")

	entry := hook.LastEntry()
	if entry == nil || entry.Level != logrus.ErrorLevel || entry.Message != "can not send message" {
		t.Errorf("entry = %+v", entry)
	}
}

func TestClient_SendMessageToTopic(t *testing.T) {
	hook := useHook(t)

	config := mocks.NewTestConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewAsyncProducer(t, config)
	producer.ExpectInputAndSucceed()

	client := NewClient()
	client.HasProducer = true
	client.TopicsPrefix = "app_"
	client.Producer = producer

	client.SendMessageToTopic("hello", "users")

	select {
	case msg := <-producer.Successes():
		if msg.Topic != "app_users" {
			t.Errorf("topic = %q", msg.Topic)
		}
		value, err := msg.Value.Encode()
		if err != nil || string(value) != "hello" {
			t.Errorf("value = %q, %v", value, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the message was not sent")
	}

	if err := producer.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	waitForLog(t, hook, "kafka message sent")
}

// blockedProducer is a producer whose input is never read.
type blockedProducer struct {
	sarama.AsyncProducer
	input  chan *sarama.ProducerMessage
	closed chan struct{}
}

func (p *blockedProducer) Input() chan<- *sarama.ProducerMessage { return p.input }

func (p *blockedProducer) AsyncClose() { close(p.closed) }

func TestClient_SendMessageToTopic_Shutdown(t *testing.T) {
	producer := &blockedProducer{input: make(chan *sarama.ProducerMessage), closed: make(chan struct{})}

	client := NewClient()
	client.HasProducer = true
	client.Producer = producer

	client.SendMessageToTopic("hello", "users")
	client.ProducerSignals <- os.Interrupt

	select {
	case <-producer.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("the producer must be closed on shutdown")
	}
}

// waitForLog waits until a message is logged. The sending is asynchronous.
func waitForLog(t *testing.T, hook *logtest.Hook, message string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, e := range hook.AllEntries() {
			if e.Message == message {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("%q is not logged", message)
}
