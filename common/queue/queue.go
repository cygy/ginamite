package queue

import (
	"github.com/cygy/ginamite/common/log"
)

// Producer sends a message to a topic. The Kafka client of the package kafka is a Producer.
type Producer interface {
	SendMessageToTopic(message, topic string)
}

// KeyedProducer is a Producer which sends the messages with the same key to the same partition, so they are consumed in
// the order they are sent. The Kafka client of the package kafka is a KeyedProducer.
type KeyedProducer interface {
	Producer
	SendKeyedMessageToTopic(key, message, topic string)
}

// Keyed is implemented by the payloads which have a key: the messages of the same key are consumed in the order they
// are sent, such as the messages of a same user.
type Keyed interface {
	PartitionKey() string
}

// Queue sends the tasks, the cache messages, the mails and the notifications to the other servers, through a Producer.
type Queue struct {
	producer Producer
}

// New returns a Queue sending its messages with the producer.
func New(producer Producer) *Queue {
	return &Queue{producer: producer}
}

// send sends the message to the topic, with the key if it is not empty and the producer is a KeyedProducer. It logs an
// error and does nothing if the queue has no producer.
func (q *Queue) send(key, message, topic string) {
	if q == nil || q.producer == nil {
		log.WithField("topic", topic).Error("unable to send the message: the queue has no producer")
		return
	}

	if keyed, ok := q.producer.(KeyedProducer); ok && key != "" {
		keyed.SendKeyedMessageToTopic(key, message, topic)
		return
	}

	q.producer.SendMessageToTopic(message, topic)
}
