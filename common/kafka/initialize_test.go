package kafka

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/cygy/ginamite/common/broker"
)

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
	client.Initialize("localhost", 9092, "app_", false, []broker.MessageConsumer{{
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
	client.Initialize("localhost", 9092, "app_", false, []broker.MessageConsumer{
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
	client.Initialize("localhost", 9092, "", false, []broker.MessageConsumer{{Topic: "users", Action: func([]byte) error { return nil }}})
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
	client.Initialize("localhost", 9092, "", false, []broker.MessageConsumer{{Topic: "users", Action: func([]byte) error { return nil }}})

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
	client.Initialize("localhost", 9092, "", false, []broker.MessageConsumer{{Topic: "users", Action: func([]byte) error { return nil }}})

	// A closed group ends the consumer.
	client.background.Wait()
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

	producer := newFakeProducer()

	var attempts atomic.Int32
	client := NewClient()
	client.MaxAttempts, client.RetryDelay = 2, time.Millisecond
	client.attachProducer(producer)
	client.Initialize("localhost", 9092, "app_", false, []broker.MessageConsumer{{
		Topic:  "mail",
		Action: func([]byte) error { attempts.Add(1); return errors.New("smtp down") },
	}})
	t.Cleanup(client.Close)

	if m := producer.next(t); m.Topic != "app_mail_dlq" || attempts.Load() != 2 {
		t.Errorf("dead letter = %+v after %d attempts", m, attempts.Load())
	}
	if stats := client.Stats(); stats.Received != 1 || stats.Retried != 1 || stats.DeadLettered != 1 || stats.Handled != 0 {
		t.Errorf("stats = %+v", stats)
	}
}

func TestClient_Close_NotInitialized(t *testing.T) {
	NewClient().Close()
}

func TestUniqueGroup(t *testing.T) {
	a, b := broker.UniqueGroup("cache"), broker.UniqueGroup("cache")

	if !strings.HasPrefix(a, "cache_") || a == b {
		t.Errorf("groups = %q, %q", a, b)
	}
}
