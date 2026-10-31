package kafka

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

// fakeSession records the messages marked as consumed.
type fakeSession struct {
	sarama.ConsumerGroupSession

	ctx    context.Context
	mu     sync.Mutex
	events *[]string
}

func (s *fakeSession) Context() context.Context { return s.ctx }

func (s *fakeSession) MarkMessage(message *sarama.ConsumerMessage, _ string) {
	if s.events != nil {
		s.mu.Lock()
		*s.events = append(*s.events, "mark:"+string(message.Value))
		s.mu.Unlock()
	}
}

// fakeClaim gives the messages of a partition.
type fakeClaim struct {
	sarama.ConsumerGroupClaim

	messages chan *sarama.ConsumerMessage
}

func (c *fakeClaim) Messages() <-chan *sarama.ConsumerMessage { return c.messages }

func TestGroupHandler_ConsumeClaim(t *testing.T) {
	t.Run("marks a message once it is handled", func(t *testing.T) {
		useHook(t)

		var events []string
		handler := groupHandler{topic: "app_users", action: func(value []byte) { events = append(events, "handle:"+string(value)) }}

		messages := make(chan *sarama.ConsumerMessage, 2)
		messages <- &sarama.ConsumerMessage{Value: []byte("a")}
		messages <- &sarama.ConsumerMessage{Value: []byte("b")}
		close(messages)

		session := &fakeSession{ctx: context.Background(), events: &events}
		if err := handler.ConsumeClaim(session, &fakeClaim{messages: messages}); err != nil {
			t.Fatal(err)
		}

		if want := []string{"handle:a", "mark:a", "handle:b", "mark:b"}; !slices.Equal(events, want) {
			t.Errorf("events = %v, want %v", events, want)
		}
	})

	t.Run("stops when the session ends", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		handler := groupHandler{action: func([]byte) { t.Error("no message to handle") }}

		done := make(chan error, 1)
		go func() {
			done <- handler.ConsumeClaim(&fakeSession{ctx: ctx}, &fakeClaim{messages: make(chan *sarama.ConsumerMessage)})
		}()
		cancel()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("ConsumeClaim() error = %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("ConsumeClaim() must stop with the session")
		}
	})
}

func TestGroupHandler_SetupAndCleanup(t *testing.T) {
	handler := groupHandler{}

	if err := handler.Setup(nil); err != nil {
		t.Errorf("Setup() error = %v", err)
	}
	if err := handler.Cleanup(nil); err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}
}
