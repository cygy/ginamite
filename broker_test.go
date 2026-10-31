package ginamite

import (
	"testing"

	"github.com/cygy/ginamite/common/broker"
)

// dedupBroker is a broker which ignores the messages delivered twice.
type dedupBroker struct {
	*broker.Memory

	deduplicator broker.Deduplicator
}

func (b *dedupBroker) UseDeduplicator(d broker.Deduplicator) { b.deduplicator = d }

func TestServer_startBroker(t *testing.T) {
	consumers := []broker.MessageConsumer{{Topic: "mail", Action: func([]byte) error { return nil }}}

	t.Run("a broker starts with the consumers", func(t *testing.T) {
		s := NewServer()
		memory := broker.NewMemory()
		s.UseBroker(memory)

		var handled bool
		if !s.startBroker([]broker.MessageConsumer{{Topic: "x", Action: func([]byte) error { handled = true; return nil }}}) {
			t.Fatal("the broker did not start")
		}
		memory.SendMessageToTopic("m", "x")
		if !handled {
			t.Error("the consumers are not started")
		}
	})

	t.Run("the broker receives the deduplicator", func(t *testing.T) {
		s := NewServer()
		b := &dedupBroker{Memory: broker.NewMemory()}
		s.UseBroker(b)
		s.startBroker(consumers)

		if b.deduplicator != s.Deduplicator || b.deduplicator == nil {
			t.Errorf("deduplicator = %v, want the one of the server", b.deduplicator)
		}
	})

	t.Run("no broker", func(t *testing.T) {
		s := NewServer()
		if s.startBroker(consumers) {
			t.Error("nothing to start without a broker")
		}
	})
}
