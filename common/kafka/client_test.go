package kafka

import (
	"context"
	"errors"
	"maps"
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
	if c.MaxAttempts != DefaultMaxAttempts || c.RetryDelay != DefaultRetryDelay {
		t.Errorf("attempts = %d, delay = %v, want the defaults", c.MaxAttempts, c.RetryDelay)
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
	for _, entry := range hook.AllEntries() {
		if _, ok := entry.Data["message"]; ok {
			t.Errorf("the body is logged: %+v", entry)
		}
	}
}

// blockedProducer is a producer whose input is never read.
type blockedProducer struct {
	sarama.AsyncProducer
	input  chan *sarama.ProducerMessage
	closed chan struct{}
}

func (p *blockedProducer) Input() chan<- *sarama.ProducerMessage { return p.input }

func (p *blockedProducer) Close() error {
	close(p.closed)

	return nil
}

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

func TestClient_Close_Producer(t *testing.T) {
	producer := &blockedProducer{input: make(chan *sarama.ProducerMessage), closed: make(chan struct{})}

	client := NewClient()
	client.Producer = producer

	// The producer is closed once, whatever the number of calls.
	client.Close()
	client.Close()

	select {
	case <-producer.closed:
	default:
		t.Error("the producer must be closed")
	}
}

func TestClient_Close_ConsumerTooLong(t *testing.T) {
	hook := useHook(t)

	old := closeTimeout
	closeTimeout = 20 * time.Millisecond
	t.Cleanup(func() { closeTimeout = old })

	release := make(chan struct{})
	client := NewClient()
	client.consumers.Add(1)
	go func() {
		defer client.consumers.Done()
		<-release
	}()

	client.Close()
	close(release)

	waitForLog(t, hook, "the kafka consumers did not stop in time")
}

func TestClient_SendMessageToTopic_ProducerNotConnected(t *testing.T) {
	hook := useHook(t)

	client := NewClient()
	client.HasProducer = true
	client.SendMessageToTopic("hello", "users")

	waitForLog(t, hook, "can not send message")
}

func TestClient_sendDeadLetter(t *testing.T) {
	message := &sarama.ConsumerMessage{Topic: "app_mail", Partition: 2, Offset: 41, Key: []byte("k"), Value: []byte("value")}

	t.Run("sends the message with the reason", func(t *testing.T) {
		hook := useHook(t)

		var sent *sarama.ProducerMessage
		producer := mocks.NewAsyncProducer(t, mocks.NewTestConfig())
		producer.ExpectInputWithMessageCheckerFunctionAndSucceed(func(m *sarama.ProducerMessage) error {
			sent = m

			return nil
		})

		client := NewClient()
		client.Producer = producer

		client.sendDeadLetter(context.Background(), message, "app_mail", errors.New("boom"), 3)

		if err := producer.Close(); err != nil {
			t.Fatal(err)
		}
		if sent == nil || sent.Topic != "app_mail_dlq" {
			t.Fatalf("sent = %+v, want a message of the topic app_mail_dlq", sent)
		}
		if value, _ := sent.Value.Encode(); string(value) != "value" {
			t.Errorf("value = %q", value)
		}
		if key, _ := sent.Key.Encode(); string(key) != "k" {
			t.Errorf("key = %q", key)
		}

		headers := map[string]string{}
		for _, h := range sent.Headers {
			headers[string(h.Key)] = string(h.Value)
		}
		want := map[string]string{"error": "boom", "attempts": "3", "group": "app_mail", "topic": "app_mail", "partition": "2", "offset": "41"}
		if !maps.Equal(headers, want) {
			t.Errorf("headers = %v, want %v", headers, want)
		}
		waitForLog(t, hook, "kafka message dead-lettered")
	})

	t.Run("no producer", func(t *testing.T) {
		hook := useHook(t)

		NewClient().sendDeadLetter(context.Background(), message, "app_mail", errors.New("boom"), 3)

		waitForLog(t, hook, "kafka message dropped: no producer for the dead letters")
	})

	t.Run("interrupted", func(t *testing.T) {
		hook := useHook(t)

		client := NewClient()
		client.Producer = &blockedProducer{input: make(chan *sarama.ProducerMessage), closed: make(chan struct{})}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		client.sendDeadLetter(ctx, message, "app_mail", errors.New("boom"), 3)

		for _, entry := range hook.AllEntries() {
			if entry.Message == "kafka message dead-lettered" {
				t.Error("the message is not sent when the client is stopped")
			}
		}
	})
}
