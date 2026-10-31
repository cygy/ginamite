package kafka

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
)

// newMockBroker returns a broker answering the metadata, offset and fetch requests of a single-partition topic.
func newMockBroker(t *testing.T, topic string) *sarama.MockBroker {
	t.Helper()

	broker := sarama.NewMockBroker(t, 1)
	t.Cleanup(broker.Close)

	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"ApiVersionsRequest": sarama.NewMockApiVersionsResponse(t).SetApiKeys([]sarama.ApiVersionsResponseKey{
			{ApiKey: 0, MinVersion: 0, MaxVersion: 8},  // Produce
			{ApiKey: 1, MinVersion: 0, MaxVersion: 11}, // Fetch
			{ApiKey: 2, MinVersion: 0, MaxVersion: 5},  // ListOffsets
			{ApiKey: 3, MinVersion: 0, MaxVersion: 8},  // Metadata
		}),
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetLeader(topic, 0, broker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset(topic, 0, sarama.OffsetNewest, 0).
			SetOffset(topic, 0, sarama.OffsetOldest, 0),
		"FetchRequest": sarama.NewMockFetchResponse(t, 1).
			SetMessage(topic, 0, 0, sarama.StringEncoder("consumed")),
	})

	return broker
}

func splitAddress(t *testing.T, addr string) (string, int) {
	t.Helper()

	host, port, found := strings.Cut(addr, ":")
	if !found {
		t.Fatalf("invalid address %q", addr)
	}

	n, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	return host, n
}

// fakeGroup is a consumer group which calls consume each time it is asked to consume.
type fakeGroup struct {
	sarama.ConsumerGroup

	consume func(ctx context.Context, handler sarama.ConsumerGroupHandler) error
	errs    chan error
	closed  chan struct{}
}

func newFakeGroup(consume func(ctx context.Context, handler sarama.ConsumerGroupHandler) error) *fakeGroup {
	return &fakeGroup{consume: consume, errs: make(chan error, 1), closed: make(chan struct{})}
}

func (g *fakeGroup) Consume(ctx context.Context, _ []string, handler sarama.ConsumerGroupHandler) error {
	return g.consume(ctx, handler)
}

func (g *fakeGroup) Errors() <-chan error { return g.errs }

func (g *fakeGroup) Close() error {
	close(g.errs)
	close(g.closed)

	return nil
}

// useGroup replaces the creation of the consumer groups, and returns the channel of the names of the created groups.
func useGroup(t *testing.T, create func(name string) (sarama.ConsumerGroup, error)) chan string {
	t.Helper()

	names := make(chan string, 10)
	oldNew, oldDelay := newConsumerGroup, retryDelay
	newConsumerGroup = func(_ []string, name string, _ *sarama.Config) (sarama.ConsumerGroup, error) {
		names <- name

		return create(name)
	}
	retryDelay = time.Millisecond
	t.Cleanup(func() { newConsumerGroup, retryDelay = oldNew, oldDelay })

	return names
}

// blockUntilDone is the Consume of a group with no message.
func blockUntilDone(ctx context.Context, _ sarama.ConsumerGroupHandler) error {
	<-ctx.Done()

	return nil
}

func TestClient_Initialize_Consumer(t *testing.T) {
	hook := useHook(t)

	var calls atomic.Int32
	group := newFakeGroup(func(ctx context.Context, handler sarama.ConsumerGroupHandler) error {
		if calls.Add(1) > 1 {
			return blockUntilDone(ctx, handler)
		}

		messages := make(chan *sarama.ConsumerMessage, 1)
		messages <- &sarama.ConsumerMessage{Value: []byte("consumed")}
		close(messages)

		return handler.ConsumeClaim(&fakeSession{ctx: ctx}, &fakeClaim{messages: messages})
	})
	names := useGroup(t, func(string) (sarama.ConsumerGroup, error) { return group, nil })

	received := make(chan []byte, 1)
	client := NewClient()
	client.Initialize("localhost", 9092, "app_", false, []MessageConsumer{{
		Topic:  "users",
		Action: func(value []byte) error { received <- value; return nil },
	}})

	if client.HasProducer || client.TopicsPrefix != "app_" {
		t.Errorf("client = %+v", client)
	}

	select {
	case value := <-received:
		if string(value) != "consumed" {
			t.Errorf("value = %q", value)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the message was not consumed")
	}
	if name := <-names; name != "app_users" {
		t.Errorf("group = %q, want the prefixed topic", name)
	}
	waitForLog(t, hook, "kafka message received")

	// The consumer leaves its group when the client is closed.
	client.Close()
	select {
	case <-group.closed:
	default:
		t.Error("the group must be closed")
	}
}

func TestClient_Initialize_ConsumerGroup(t *testing.T) {
	names := useGroup(t, func(string) (sarama.ConsumerGroup, error) { return newFakeGroup(blockUntilDone), nil })

	client := NewClient()
	client.Initialize("localhost", 9092, "app_", false, []MessageConsumer{
		{Topic: "users", Group: "workers", Action: func([]byte) error { return nil }},
		{Topic: "mail", Action: func([]byte) error { return nil }},
	})
	t.Cleanup(client.Close)

	got := []string{<-names, <-names}
	slices.Sort(got)
	if want := []string{"app_mail", "app_workers"}; !slices.Equal(got, want) {
		t.Errorf("groups = %v, want %v", got, want)
	}
}

func TestClient_consume_RetriesToConnect(t *testing.T) {
	hook := useHook(t)

	var calls atomic.Int32
	group := newFakeGroup(blockUntilDone)
	names := useGroup(t, func(string) (sarama.ConsumerGroup, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("no broker")
		}

		return group, nil
	})

	client := NewClient()
	client.Initialize("localhost", 9092, "", false, []MessageConsumer{{Topic: "users", Action: func([]byte) error { return nil }}})
	t.Cleanup(client.Close)

	<-names
	<-names
	waitForLog(t, hook, "unable to connect to kafka")
	waitForLog(t, hook, "connected to kafka")
}

func TestClient_consume_StopsWhileConnecting(t *testing.T) {
	useGroup(t, func(string) (sarama.ConsumerGroup, error) { return nil, errors.New("no broker") })
	retryDelay = time.Hour

	client := NewClient()
	client.Initialize("localhost", 9092, "", false, []MessageConsumer{{Topic: "users", Action: func([]byte) error { return nil }}})

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

func TestClient_consume_Errors(t *testing.T) {
	hook := useHook(t)

	var calls atomic.Int32
	var group *fakeGroup
	group = newFakeGroup(func(ctx context.Context, handler sarama.ConsumerGroupHandler) error {
		switch calls.Add(1) {
		case 1:
			group.errs <- errors.New("rebalance error")
			return errors.New("consume failed")
		default:
			return sarama.ErrClosedConsumerGroup
		}
	})
	useGroup(t, func(string) (sarama.ConsumerGroup, error) { return group, nil })

	client := NewClient()
	client.Initialize("localhost", 9092, "", false, []MessageConsumer{{Topic: "users", Action: func([]byte) error { return nil }}})

	// A closed group ends the consumer.
	client.consumers.Wait()
	client.Close()

	waitForLog(t, hook, "kafka consume error")
	waitForLog(t, hook, "kafka consumer error")
	if calls.Load() != 2 {
		t.Errorf("Consume() called %d times, want a retry after the error and no more after the group was closed", calls.Load())
	}
}

func TestClient_Initialize_DeadLetter(t *testing.T) {
	useHook(t)

	var calls atomic.Int32
	group := newFakeGroup(func(ctx context.Context, handler sarama.ConsumerGroupHandler) error {
		if calls.Add(1) > 1 {
			return blockUntilDone(ctx, handler)
		}

		messages := make(chan *sarama.ConsumerMessage, 1)
		messages <- &sarama.ConsumerMessage{Topic: "app_mail", Value: []byte("bad")}
		close(messages)

		return handler.ConsumeClaim(&fakeSession{ctx: ctx}, &fakeClaim{messages: messages})
	})
	useGroup(t, func(string) (sarama.ConsumerGroup, error) { return group, nil })

	producer := mocks.NewAsyncProducer(t, mocks.NewTestConfig())
	dead := make(chan *sarama.ProducerMessage, 1)
	producer.ExpectInputWithMessageCheckerFunctionAndSucceed(func(m *sarama.ProducerMessage) error {
		dead <- m

		return nil
	})

	var attempts atomic.Int32
	client := NewClient()
	client.MaxAttempts, client.RetryDelay = 2, time.Millisecond
	client.Producer = producer
	client.Initialize("localhost", 9092, "app_", false, []MessageConsumer{{
		Topic:  "mail",
		Action: func([]byte) error { attempts.Add(1); return errors.New("smtp down") },
	}})
	t.Cleanup(client.Close)

	select {
	case m := <-dead:
		if m.Topic != "app_mail_dlq" || attempts.Load() != 2 {
			t.Errorf("dead letter = %+v after %d attempts", m, attempts.Load())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the message was not dead-lettered")
	}
}

func TestClient_Close_NotInitialized(t *testing.T) {
	NewClient().Close()
}

func TestUniqueGroup(t *testing.T) {
	a, b := UniqueGroup("cache"), UniqueGroup("cache")

	if !strings.HasPrefix(a, "cache_") || a == b {
		t.Errorf("groups = %q, %q", a, b)
	}
}

func TestClient_Initialize_Producer(t *testing.T) {
	hook := useHook(t)
	broker := newMockBroker(t, "app_users")
	host, port := splitAddress(t, broker.Addr())

	client := NewClient()
	client.Initialize(host, port, "app_", true, nil)

	if !client.HasProducer {
		t.Fatal("the client must be a producer")
	}

	waitForLog(t, hook, "connected to kafka")
}
