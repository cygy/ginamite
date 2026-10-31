package kafka

import (
	"context"
	"errors"
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
	events *[]string
}

func (s *fakeSession) Context() context.Context { return s.ctx }

func (s *fakeSession) MarkMessage(message *sarama.ConsumerMessage, _ string) {
	if s.events != nil {
		*s.events = append(*s.events, "mark:"+string(message.Value))
	}
}

// fakeClaim gives the messages of a partition.
type fakeClaim struct {
	sarama.ConsumerGroupClaim

	messages chan *sarama.ConsumerMessage
}

func (c *fakeClaim) Messages() <-chan *sarama.ConsumerMessage { return c.messages }

// deadLetters collects the dead letters of a handler.
type deadLetters struct {
	mu       sync.Mutex
	messages []string
	errs     []error
	attempts []int
}

func (d *deadLetters) add(message *sarama.ConsumerMessage, err error, attempts int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.messages = append(d.messages, string(message.Value))
	d.errs = append(d.errs, err)
	d.attempts = append(d.attempts, attempts)
}

func TestGroupHandler_ConsumeClaim(t *testing.T) {
	t.Run("marks a message once it is handled", func(t *testing.T) {
		useHook(t)

		var events []string
		handler := groupHandler{topic: "app_users", attempts: 1, action: func(value []byte) error {
			events = append(events, "handle:"+string(value))

			return nil
		}}

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

	t.Run("dead-letters a failed message, marks it and goes on", func(t *testing.T) {
		useHook(t)

		var events []string
		letters := &deadLetters{}
		handler := groupHandler{attempts: 2, delay: time.Millisecond, deadLetter: letters.add, action: func(value []byte) error {
			events = append(events, "handle:"+string(value))
			if string(value) == "bad" {
				return errors.New("boom")
			}

			return nil
		}}

		messages := make(chan *sarama.ConsumerMessage, 2)
		messages <- &sarama.ConsumerMessage{Value: []byte("bad")}
		messages <- &sarama.ConsumerMessage{Value: []byte("good")}
		close(messages)

		session := &fakeSession{ctx: context.Background(), events: &events}
		if err := handler.ConsumeClaim(session, &fakeClaim{messages: messages}); err != nil {
			t.Fatal(err)
		}

		want := []string{"handle:bad", "handle:bad", "mark:bad", "handle:good", "mark:good"}
		if !slices.Equal(events, want) {
			t.Errorf("events = %v, want %v", events, want)
		}
		if !slices.Equal(letters.messages, []string{"bad"}) || letters.attempts[0] != 2 {
			t.Errorf("dead letters = %v after %v attempts", letters.messages, letters.attempts)
		}
	})

	t.Run("stops when the session ends", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		handler := groupHandler{action: func([]byte) error {
			t.Error("no message to handle")

			return nil
		}}

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

	t.Run("does not mark a message interrupted while it waits for a retry", func(t *testing.T) {
		useHook(t)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var events []string
		handler := groupHandler{attempts: 3, delay: time.Hour, deadLetter: func(*sarama.ConsumerMessage, error, int) {
			t.Error("an interrupted message must not be dead-lettered")
		}, action: func([]byte) error {
			cancel()

			return errors.New("boom")
		}}

		messages := make(chan *sarama.ConsumerMessage, 1)
		messages <- &sarama.ConsumerMessage{Value: []byte("a")}

		if err := handler.ConsumeClaim(&fakeSession{ctx: ctx, events: &events}, &fakeClaim{messages: messages}); err != nil {
			t.Fatal(err)
		}
		if len(events) != 0 {
			t.Errorf("events = %v, the message must be handled again by the group", events)
		}
	})
}

func TestGroupHandler_handle(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name         string
		attempts     int
		failures     int   // the number of attempts failing before one succeeds, -1 for always
		err          error // the error of the failing attempts
		wantCalls    int
		wantDeadLett bool
	}{
		{"succeeds at once", 3, 0, boom, 1, false},
		{"succeeds after retries", 3, 2, boom, 3, false},
		{"fails every attempt", 3, -1, boom, 3, true},
		{"a permanent error is not retried", 3, -1, Permanent(boom), 1, true},
		{"one attempt at least", 0, -1, boom, 1, true},
		{"a panic is retried", 2, -1, nil, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useHook(t)

			calls := 0
			letters := &deadLetters{}
			handler := groupHandler{attempts: tt.attempts, delay: time.Millisecond, deadLetter: letters.add, action: func([]byte) error {
				calls++
				if tt.failures >= 0 && calls > tt.failures {
					return nil
				}
				if tt.err == nil {
					panic("kaboom")
				}

				return tt.err
			}}

			if !handler.handle(context.Background(), &sarama.ConsumerMessage{Value: []byte("a")}) {
				t.Error("handle() = false")
			}

			if calls != tt.wantCalls {
				t.Errorf("attempts = %d, want %d", calls, tt.wantCalls)
			}
			if (len(letters.messages) == 1) != tt.wantDeadLett {
				t.Fatalf("dead letters = %v, want one: %v", letters.messages, tt.wantDeadLett)
			}
			if tt.wantDeadLett && letters.attempts[0] != tt.wantCalls {
				t.Errorf("dead-lettered after %d attempts, want %d", letters.attempts[0], tt.wantCalls)
			}
		})
	}

	t.Run("the error of a panic is kept", func(t *testing.T) {
		useHook(t)

		letters := &deadLetters{}
		handler := groupHandler{attempts: 1, deadLetter: letters.add, action: func([]byte) error { panic("kaboom") }}
		handler.handle(context.Background(), &sarama.ConsumerMessage{Value: []byte("a")})

		if len(letters.errs) != 1 || letters.errs[0].Error() != "panic: kaboom" {
			t.Errorf("errors = %v", letters.errs)
		}
	})

	t.Run("a message interrupted during its last attempt is not dead-lettered", func(t *testing.T) {
		useHook(t)

		ctx, cancel := context.WithCancel(context.Background())
		handler := groupHandler{attempts: 1, deadLetter: func(*sarama.ConsumerMessage, error, int) {
			t.Error("an interrupted message must not be dead-lettered")
		}, action: func([]byte) error {
			cancel()

			return boom
		}}

		if handler.handle(ctx, &sarama.ConsumerMessage{Value: []byte("a")}) {
			t.Error("handle() = true, want false: the message must be handled again")
		}
	})
}

func TestGroupHandler_backoff(t *testing.T) {
	handler := groupHandler{delay: time.Second}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{6, 32 * time.Second},
		{7, maxRetryDelay},
		{100, maxRetryDelay},
	}
	for _, tt := range tests {
		if got := handler.backoff(tt.attempt); got != tt.want {
			t.Errorf("backoff(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}

	if got := (groupHandler{delay: time.Hour}).backoff(1); got != maxRetryDelay {
		t.Errorf("backoff() = %v, want it capped to %v", got, maxRetryDelay)
	}
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
