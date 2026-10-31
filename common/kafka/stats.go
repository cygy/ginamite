package kafka

import (
	"fmt"
	"maps"
	"sync"
	"sync/atomic"
)

// Stats are the counters of a Client since it started.
type Stats struct {
	// Received is the number of messages consumed, Handled the number handled successfully, Retried the number of
	// attempts after a failure, and DeadLettered the number of messages which could not be handled.
	Received, Handled, Retried, DeadLettered uint64

	// Sent is the number of messages acknowledged by Kafka, SendErrors the number it refused, and Dropped the number
	// lost because the client is closed or the queue of the messages to send is full.
	Sent, SendErrors, Dropped uint64

	// Lag is the number of messages waiting to be consumed, by "topic/partition", as seen when the last message of
	// the partition was handled.
	Lag map[string]int64
}

// stats holds the counters. The methods accept a nil receiver.
type stats struct {
	received, handled, retried, deadLettered atomic.Uint64
	sent, sendErrors, dropped                atomic.Uint64

	mu  sync.Mutex
	lag map[string]int64
}

func (s *stats) countReceived() {
	if s != nil {
		s.received.Add(1)
	}
}

func (s *stats) countHandled() {
	if s != nil {
		s.handled.Add(1)
	}
}

func (s *stats) countRetried() {
	if s != nil {
		s.retried.Add(1)
	}
}

func (s *stats) countDeadLettered() {
	if s != nil {
		s.deadLettered.Add(1)
	}
}

func (s *stats) setLag(topic string, partition int32, lag int64) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lag == nil {
		s.lag = map[string]int64{}
	}
	s.lag[fmt.Sprintf("%s/%d", topic, partition)] = lag
}

func (s *stats) snapshot() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	return Stats{
		Received:     s.received.Load(),
		Handled:      s.handled.Load(),
		Retried:      s.retried.Load(),
		DeadLettered: s.deadLettered.Load(),
		Sent:         s.sent.Load(),
		SendErrors:   s.sendErrors.Load(),
		Dropped:      s.dropped.Load(),
		Lag:          maps.Clone(s.lag),
	}
}
