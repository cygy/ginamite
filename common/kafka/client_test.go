package kafka

import (
	"context"
	"errors"
	"maps"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// fakeProducer is a producer which gives its messages to the test.
type fakeProducer struct {
	sarama.AsyncProducer

	input     chan *sarama.ProducerMessage
	successes chan *sarama.ProducerMessage
	errs      chan *sarama.ProducerError
	closed    atomic.Int32
}

func newFakeProducer() *fakeProducer {
	return &fakeProducer{
		input:     make(chan *sarama.ProducerMessage, 10),
		successes: make(chan *sarama.ProducerMessage, 10),
		errs:      make(chan *sarama.ProducerError, 10),
	}
}

func (p *fakeProducer) Input() chan<- *sarama.ProducerMessage     { return p.input }
func (p *fakeProducer) Successes() <-chan *sarama.ProducerMessage { return p.successes }
func (p *fakeProducer) Errors() <-chan *sarama.ProducerError      { return p.errs }

func (p *fakeProducer) Close() error {
	if p.closed.Add(1) == 1 {
		close(p.successes)
		close(p.errs)
	}

	return nil
}

// producerClient returns a client whose producer is the fake one.
func producerClient(t *testing.T) (*Client, *fakeProducer) {
	t.Helper()

	producer := newFakeProducer()
	client := NewClient()
	client.HasProducer = true
	client.TopicsPrefix = "app_"
	client.attachProducer(producer)

	return client, producer
}

// next returns the next message the producer received.
func (p *fakeProducer) next(t *testing.T) *sarama.ProducerMessage {
	t.Helper()

	select {
	case m := <-p.input:
		return m
	case <-time.After(2 * time.Second):
		t.Fatal("the producer received no message")

		return nil
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient()

	if c.HasProducer || c.Producer != nil || cap(c.outbox) != outboxSize {
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

func TestDeadLetterTopic(t *testing.T) {
	if got := DeadLetterTopic("app_mail"); got != "app_mail_dlq" {
		t.Errorf("DeadLetterTopic() = %q", got)
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
	client, producer := producerClient(t)

	for _, message := range []string{"one", "two", "three"} {
		client.SendMessageToTopic(message, "users")
	}

	// The messages are given to the producer in the order they were sent.
	for _, want := range []string{"one", "two", "three"} {
		m := producer.next(t)
		value, err := m.Value.Encode()
		if m.Topic != "app_users" || err != nil || string(value) != want {
			t.Fatalf("message = %q on %q, %v, want %q on app_users", value, m.Topic, err, want)
		}

		m.Partition, m.Offset = 1, 7
		producer.successes <- m
	}

	waitForLog(t, hook, "kafka message sent")
	eventually(t, func() bool { return client.Stats().Sent == 3 })

	for _, entry := range hook.AllEntries() {
		if _, ok := entry.Data["message"]; ok || strings.Contains(entry.Message, "one") {
			t.Errorf("the body is logged: %+v", entry)
		}
	}
}

func TestClient_SendMessageToTopic_BeforeTheProducerIsConnected(t *testing.T) {
	useHook(t)

	client := NewClient()
	client.HasProducer = true
	client.SendMessageToTopic("early", "users")

	// The message waits in the outbox.
	if len(client.outbox) != 1 {
		t.Fatalf("the outbox holds %d messages", len(client.outbox))
	}

	producer := newFakeProducer()
	client.attachProducer(producer)

	if m := producer.next(t); m.Topic != "users" {
		t.Errorf("message = %+v", m)
	}
}

func TestClient_SendMessageToTopic_Full(t *testing.T) {
	hook := useHook(t)

	client := NewClient()
	client.HasProducer = true
	for range outboxSize + 2 {
		client.SendMessageToTopic("x", "users")
	}

	if client.Stats().Dropped != 2 {
		t.Errorf("dropped = %d, want 2", client.Stats().Dropped)
	}
	waitForLog(t, hook, "kafka message dropped: too many messages wait for the producer")
}

func TestClient_SendMessageToTopic_Closed(t *testing.T) {
	hook := useHook(t)
	client, _ := producerClient(t)

	client.Close()
	client.SendMessageToTopic("late", "users")

	if client.Stats().Dropped != 1 {
		t.Errorf("dropped = %d, want 1", client.Stats().Dropped)
	}
	waitForLog(t, hook, "kafka message dropped: the client is closed")
}

func TestClient_producerErrors(t *testing.T) {
	hook := useHook(t)
	client, producer := producerClient(t)

	producer.errs <- &sarama.ProducerError{Msg: &sarama.ProducerMessage{Topic: "app_users"}, Err: errors.New("broker down")}

	waitForLog(t, hook, "kafka send message error")
	eventually(t, func() bool { return client.Stats().SendErrors == 1 })
}

func TestClient_Close_Producer(t *testing.T) {
	client, producer := producerClient(t)

	client.SendMessageToTopic("last", "users")

	// The waiting messages are given to the producer before it is closed, and it is closed once.
	client.Close()
	client.Close()

	if got := len(producer.input); got != 1 {
		t.Errorf("the producer received %d messages, want 1", got)
	}
	if producer.closed.Load() != 1 {
		t.Errorf("the producer was closed %d times, want once", producer.closed.Load())
	}
}

func TestClient_Close_WithoutProducer(t *testing.T) {
	hook := useHook(t)

	client := NewClient()
	client.HasProducer = true
	client.SendMessageToTopic("lost", "users")
	client.Close()

	if client.Stats().Dropped != 1 {
		t.Errorf("dropped = %d, want 1", client.Stats().Dropped)
	}
	waitForLog(t, hook, "kafka messages dropped: the producer was never connected")
}

func TestClient_Close_ConsumerTooLong(t *testing.T) {
	hook := useHook(t)

	old := closeTimeout
	closeTimeout = 20 * time.Millisecond
	t.Cleanup(func() { closeTimeout = old })

	release := make(chan struct{})
	client := NewClient()
	client.background.Add(1)
	go func() {
		defer client.background.Done()
		<-release
	}()

	client.Close()
	close(release)

	waitForLog(t, hook, "the kafka consumers did not stop in time")
}

func TestClient_Close_ProducerTooLong(t *testing.T) {
	hook := useHook(t)

	old := closeTimeout
	closeTimeout = 20 * time.Millisecond
	t.Cleanup(func() { closeTimeout = old })

	client := NewClient()
	client.HasProducer = true
	client.attachProducer(&fakeProducer{input: make(chan *sarama.ProducerMessage), successes: make(chan *sarama.ProducerMessage), errs: make(chan *sarama.ProducerError)})
	client.SendMessageToTopic("stuck", "users")

	client.Close()

	waitForLog(t, hook, "the kafka producer did not send the waiting messages in time")
}

func TestClient_connectProducer(t *testing.T) {
	hook := useHook(t)

	oldNew, oldDelay := newAsyncProducer, retryDelay
	retryDelay = time.Millisecond
	t.Cleanup(func() { newAsyncProducer, retryDelay = oldNew, oldDelay })

	var calls atomic.Int32
	producer := newFakeProducer()
	newAsyncProducer = func([]string, *sarama.Config) (sarama.AsyncProducer, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("no broker")
		}

		return producer, nil
	}

	client := NewClient()
	client.Initialize("localhost", 9092, "app_", true, nil)
	t.Cleanup(client.Close)

	waitForLog(t, hook, "unable to connect to kafka")
	waitForLog(t, hook, "connected to kafka")
	eventually(t, func() bool { return client.producer() == producer })
}

func TestClient_connectProducer_Stopped(t *testing.T) {
	oldNew, oldDelay := newAsyncProducer, retryDelay
	retryDelay = time.Hour
	t.Cleanup(func() { newAsyncProducer, retryDelay = oldNew, oldDelay })
	newAsyncProducer = func([]string, *sarama.Config) (sarama.AsyncProducer, error) { return nil, errors.New("no broker") }

	client := NewClient()
	client.Initialize("localhost", 9092, "", true, nil)

	done := make(chan struct{})
	go func() {
		client.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Close() must not wait for the next attempt")
	}
}

func TestClient_logStats(t *testing.T) {
	hook := useHook(t)

	client := NewClient()
	client.Options.StatsInterval = 5 * time.Millisecond
	client.stats.received.Add(3)
	client.stats.setLag("app_mail", 0, 4)
	client.Initialize("localhost", 9092, "app_", false, nil)
	t.Cleanup(client.Close)

	waitForLog(t, hook, "kafka stats")

	for _, entry := range hook.AllEntries() {
		if entry.Message == "kafka stats" {
			if entry.Data["received"] != uint64(3) || entry.Data["lag app_mail/0"] != int64(4) {
				t.Errorf("stats = %v", entry.Data)
			}

			return
		}
	}
}

func TestClient_Stats(t *testing.T) {
	client := NewClient()
	client.stats.setLag("app_mail", 1, 9)

	snapshot := client.Stats()
	snapshot.Lag["app_mail/1"] = 0

	if got := client.Stats().Lag; !maps.Equal(got, map[string]int64{"app_mail/1": 9}) {
		t.Errorf("lag = %v, the snapshot must be a copy", got)
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

// eventually waits until the condition is true.
func eventually(t *testing.T, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("the condition is never true")
}

func TestClient_sendDeadLetter(t *testing.T) {
	message := &sarama.ConsumerMessage{Topic: "app_mail", Partition: 2, Offset: 41, Key: []byte("k"), Value: []byte("value")}

	t.Run("sends the message with the reason", func(t *testing.T) {
		hook := useHook(t)

		client := NewClient()
		producer := newFakeProducer()
		client.Producer = producer

		client.sendDeadLetter(context.Background(), message, "app_mail", errors.New("boom"), 3)

		sent := producer.next(t)
		if sent.Topic != "app_mail_dlq" {
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

		client := NewClient()
		client.sendDeadLetter(context.Background(), message, "app_mail", errors.New("boom"), 3)

		waitForLog(t, hook, "kafka message dropped: no producer for the dead letters")
		if client.Stats().Dropped != 1 {
			t.Errorf("dropped = %d", client.Stats().Dropped)
		}
	})

	t.Run("interrupted", func(t *testing.T) {
		hook := useHook(t)

		client := NewClient()
		client.Producer = &fakeProducer{input: make(chan *sarama.ProducerMessage)}

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

func TestClient_UseProducer(t *testing.T) {
	client := NewClient()
	producer := newFakeProducer()
	client.UseProducer(producer)

	client.SendMessageToTopic("hello", "users")

	if !client.HasProducer || producer.next(t).Topic != "users" {
		t.Error("the producer of the caller must receive the messages")
	}
}
