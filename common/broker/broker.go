// Package broker defines what the servers need from a message broker: a Producer to send the messages, and consumers
// to handle the messages of the topics. The package kafka implements it with Apache Kafka, and Memory implements it
// for the tests. Another broker is used by implementing Broker, and giving it to the server before it starts.
package broker

import (
	"errors"

	"github.com/cygy/ginamite/common/random"
)

// Producer sends a message to a topic.
type Producer interface {
	SendMessageToTopic(message, topic string)
}

// KeyedProducer is a Producer which delivers the messages with the same key in the order they are sent, such as by
// sending them to the same partition.
type KeyedProducer interface {
	Producer
	SendKeyedMessageToTopic(key, message, topic string)
}

// Broker connects the servers to a message broker.
type Broker interface {
	KeyedProducer

	// Start connects to the broker in the background and starts the consumers: it returns without waiting.
	Start(consumers []MessageConsumer)
	// Close stops the consumers, once the message they handle is done, and sends the messages not sent yet.
	Close()
}

// MessageConsumer consumes the messages of a topic and executes Action on each one.
//
// The consumers of the same Group share the messages of the topic: each message is handled by one of them, so the
// workers sharing a group share the load. The consumers of different groups all receive every message, see UniqueGroup.
type MessageConsumer struct {
	Topic string
	// Group is the name of the consumer group, prefixed like the topics. It is the name of the topic if empty.
	Group string
	// Action handles a message and returns an error if it failed. A message is handled again by the group if the
	// process stops before it is acknowledged, so Action must tolerate a message handled twice.
	//
	// The broker handles a failed message again, then sets it aside as a dead letter: the number of attempts and
	// the place of the dead letters depend on the broker. An error marked with Permanent is not retried. A panic is
	// recovered and handled like an error.
	Action func(value []byte) error
	// Idempotent says Action gives the same result when a message is handled twice, such as the invalidation of a
	// cache: the broker does not spend a Deduplicator on its messages.
	Idempotent bool
}

// UniqueGroup returns the name of a group no other process uses, so its consumer receives every message of the topic:
// use it for the messages every instance of a server must handle, such as the invalidations of a cache.
func UniqueGroup(name string) string {
	return name + "_" + random.String(8)
}

// permanentError is an error handling the message again can not fix.
type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }

func (e permanentError) Unwrap() error { return e.err }

// Permanent marks the error of an action as one a retry can not fix, such as a message which can not be decoded: the
// message is dead-lettered at once, instead of being tried again. It returns nil if err is nil.
func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return permanentError{err}
}

// IsPermanent reports whether the error, or one it wraps, is marked with Permanent.
func IsPermanent(err error) bool {
	return errors.As(err, &permanentError{})
}
