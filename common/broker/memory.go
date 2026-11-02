package broker

import (
	"fmt"
	"sync"
)

// Sent is a message sent to a Memory broker.
type Sent struct {
	Topic, Key, Message string
}

// Memory is a Broker keeping the messages in memory, for the tests: a message is recorded, then handled at once, in the
// goroutine of the sender, by the consumers of its topic. Nothing is retried nor dead-lettered: the errors of the
// consumers are recorded.
type Memory struct {
	mu        sync.Mutex
	consumers map[string][]MessageConsumer
	sent      []Sent
	errs      []error
	closed    bool
}

// NewMemory returns an empty Memory broker.
func NewMemory() *Memory {
	return &Memory{consumers: map[string][]MessageConsumer{}}
}

// Start registers the consumers.
func (m *Memory) Start(consumers []MessageConsumer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range consumers {
		m.consumers[c.Topic] = append(m.consumers[c.Topic], c)
	}
}

// Close stops delivering the messages: the ones sent later are ignored.
func (m *Memory) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
}

// SendMessageToTopic records the message and delivers it, see SendKeyedMessageToTopic.
func (m *Memory) SendMessageToTopic(message, topic string) {
	m.SendKeyedMessageToTopic("", message, topic)
}

// SendKeyedMessageToTopic records the message and calls the consumers of the topic with it. A consumer of a group
// shared with another one receives the message too: the groups are not simulated.
func (m *Memory) SendKeyedMessageToTopic(key, message, topic string) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.sent = append(m.sent, Sent{Topic: topic, Key: key, Message: message})
	consumers := append([]MessageConsumer(nil), m.consumers[topic]...)
	m.mu.Unlock()

	for _, c := range consumers {
		if err := m.deliver(c, message); err != nil {
			m.mu.Lock()
			m.errs = append(m.errs, err)
			m.mu.Unlock()
		}
	}
}

// deliver calls the action of the consumer, and turns its panic into an error.
func (m *Memory) deliver(c MessageConsumer, message string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return c.Action([]byte(message))
}

// Sent returns the messages sent so far, in order.
func (m *Memory) Sent() []Sent {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]Sent(nil), m.sent...)
}

// Errors returns the errors returned by the consumers so far.
func (m *Memory) Errors() []error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]error(nil), m.errs...)
}
